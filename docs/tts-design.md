# TTS 合成设计文档

本文档整理 VoStory TTS 合成链路的设计思路，涵盖发音词典、声音配置、情绪体系及与 TTS 引擎的关联。

---

## 一、整体合成流程

```
┌─────────────┐    ┌──────────────┐    ┌───────────────┐
│ ScriptSegment│───→│ 发音词典处理  │───→│  文本（已纠音）│
│ .content     │    │ (有效词典合并) │    └───────┬───────┘
└──────┬───────┘    └──────────────┘            │
       │                                        │
       │ character_id                            │
       ▼                                        │
┌─────────────┐    ┌──────────────┐             │
│  Character   │───→│ VoiceProfile │             │
│              │    │ .ref_audio   │             │
└──────────────┘    │ .tts_provider│             │
                    │ .tts_params  │             │
                    └──────┬───────┘             │
                           │                     │
                    ┌──────▼───────┐             │
                    │ VoiceEmotion │             │
                    │ (按情绪+强度  │             │
                    │  查参考音频)  │             │
                    └──────┬───────┘             │
                           │                     │
                    ┌──────▼─────────────────────▼──┐
                    │        TTS 适配层              │
                    │  (统一接口，多引擎切换)          │
                    └──────────┬────────────────────┘
                               │
                        ┌──────▼───────┐
                        │  AudioClip   │
                        │ (生成结果)    │
                        └──────────────┘
```

合成一个脚本片段时，分两条线并行准备：

- **文本侧**：片段原文 → 经发音词典纠音 → 得到可发音文本
- **音频侧**：片段角色 → 声音配置 → 按情绪查参考音频 → 确定 TTS 引擎和参数

两侧汇合后交给 TTS 适配层，生成 AudioClip。

---

## 二、发音词典（PronunciationDict）

### 解决的问题

TTS 引擎遇到多音字、专有名词、生僻字时经常读错。发音词典是一个"纠错表"，在合成前对文本进行预处理。

### 数据模型

| 字段 | 说明 |
|------|------|
| `word` | 原始词（如"重庆"、"单于"、"乐正绫"） |
| `phoneme` | 正确发音标注（如"chóng qìng"、"chán yú"） |
| `project_id` | NULL = 工作空间全局词条，有值 = 项目级词条 |
| `workspace_id` | 所属工作空间 |

### 两级覆盖机制

```
全局词典（workspace 级，project_id IS NULL）
    └── 项目词典（project 级，同词覆盖全局）
         └── 合并后的「有效词典」→ 传给 TTS 引擎
```

`FindEffective` 接口逻辑：

1. 取该项目下所有项目级词条
2. 取该工作空间下全局词条
3. 同词以项目级为准，全局级被覆盖

### 与 TTS 的对接方式

合成时将有效词典转换为 IndexTTS2 支持的拼音标注格式，混入文本中传给 TTS 引擎。IndexTTS2 原生支持中文混合拼音标注。

### 当前状态

- ✅ 后端 CRUD + 有效词典合并接口
- ✅ 前端管理页面
- ⬜ 合成时消费有效词典（第 3 阶段）

---

## 三、声音配置体系

### 三层声音架构

```
VoiceAsset（工作空间级，全局声音资产库，跨项目复用）
    └── VoiceProfile（项目级，声音配置，绑定到角色）
         └── VoiceEmotion（情绪级，不同情绪+强度的参考音频）
```

### VoiceAsset — 声音资产

工作空间级别的音色库，可被多个项目的 VoiceProfile 引用。

| 字段 | 说明 |
|------|------|
| `workspace_id` | 所属工作空间 |
| `name` | 音色名称 |
| `gender` | male / female / unknown |
| `reference_audio_url` | 默认参考音频 |
| `reference_text` | 参考音频对应文本 |
| `tts_provider_id` | 关联 TTS 提供商 |
| `tags` | 标签（JSON 数组） |

### VoiceProfile — 声音配置

项目级别，为角色定义"用什么声音说话"。

| 字段 | 说明 |
|------|------|
| `project_id` | 所属项目 |
| `voice_asset_id` | 可选，引用全局声音资产 |
| `name` | 配置名称（如"男主-低沉版"） |
| `reference_audio_url` | 项目级参考音频（可覆盖全局资产） |
| `reference_text` | 参考音频对应文本 |
| `tts_provider_id` | 可选，指定 TTS 引擎（覆盖项目默认） |
| `tts_params` | JSON，引擎特定参数（语速、音调等） |

### VoiceEmotion — 情绪参考音频

为同一个声音配置提供不同情绪下的参考音频样本。

| 字段 | 说明 |
|------|------|
| `voice_profile_id` | 所属声音配置 |
| `emotion_type` | 情绪类型（neutral / happy / sad / angry / fear / surprise / disgust） |
| `emotion_strength` | 强度（light / medium / strong） |
| `reference_audio_url` | 该情绪的参考音频 |
| `reference_text` | 参考文本 |

唯一约束：`(voice_profile_id, emotion_type, emotion_strength)`，即同一配置下每种情绪+强度组合只有一条参考音频。

### 角色与声音配置的关联

```
Character.voice_profile_id → VoiceProfile
```

角色表通过 `voice_profile_id` 外键绑定声音配置。

### 当前状态

- ✅ VoiceProfile — 后端 CRUD + 前端管理页面（表单未暴露 tts_provider_id / tts_params）
- ✅ VoiceAsset — 模型已建 + 迁移
- ⬜ VoiceEmotion — 仅模型和迁移，无 CRUD / 前端页面
- ⬜ 角色绑定声音配置 — 后端 API 已支持，前端编辑页未做绑定 UI

---

## 四、情绪体系设计

### 双维度：类型 × 强度

| 维度 | 可选值 | 说明 |
|------|--------|------|
| emotion_type | neutral, happy, sad, angry, fear, surprise, disgust | 7 种基础情绪 |
| emotion_strength | light, medium, strong | 3 档强度 |

总计 7 × 3 = 21 种组合。

### 三档强度的合理性

三档强度（light / medium / strong）对 MVP 阶段足够，原因：

1. **主流 TTS 引擎的情绪控制靠参考音频驱动**，而非精确数值参数。三档强度本质是"参考音频选择器"——选择不同强度的参考音频样本。
2. **每种情绪最多 3 条参考音频**，实际制作中是合理的工作量。
3. **人耳对情绪强度的感知本身是模糊的**，过细的分级增加标注成本，TTS 引擎也无法精确还原。

### 后续扩展方向

如果三档不够用，可向后兼容地扩展：

| 场景 | 方案 |
|------|------|
| 需要更细粒度 | 增加 `very_light` / `very_strong`，变成五档 |
| 引擎支持连续值 | TTS 适配层做映射：light→0.3, medium→0.6, strong→0.9 |
| SSML 精细控制 | 语速/音量/音调通过 `tts_params` 独立控制，不混入情绪强度 |

---

## 五、TTS 合成时的查找链路

合成单个脚本片段的完整流程：

```
1. 输入：ScriptSegment
   ├── content          → 原始文本
   ├── character_id     → 角色 ID
   ├── emotion_type     → 情绪类型
   └── emotion_strength → 情绪强度

2. 文本处理：
   ├── getEffectivePronunciationDict(workspace_id, project_id)
   └── 按引擎协议将词典应用到文本

3. 声音查找：
   ├── Character → voice_profile_id → VoiceProfile
   ├── VoiceEmotion WHERE
   │     voice_profile_id = X
   │     AND emotion_type = 片段.emotion_type
   │     AND emotion_strength = 片段.emotion_strength
   ├── 找到 → 用情绪参考音频
   └── 未找到 → 回退到 VoiceProfile 默认参考音频

4. 引擎选择：
   ├── VoiceProfile.tts_provider_id（优先）
   └── Project 默认 tts_provider_id（兜底）

5. 调用 TTS 适配层：
   ├── text: 纠音后的文本
   ├── reference_audio: 参考音频 URL
   ├── reference_text: 参考文本
   └── tts_params: 引擎特定参数

6. 输出：AudioClip
   ├── audio_url, duration, file_size, format
   ├── voice_profile_id, emotion_type, emotion_strength
   └── tts_provider_id, version, is_current
```

---

## 六、TTS 引擎（IndexTTS2）

语音合成基于 IndexTTS2，需自行部署（VoStory 提供 Docker 镜像）。Go 后端通过 HTTP 协议调用：

```text
IndexTTS2 HTTP 端点：
    POST /v2/synthesize    → 合成语音（传入文本、参考音频、情绪向量）
    GET  /v1/check/audio   → 检查参考音频是否存在
    POST /v1/upload_audio  → 上传参考音频
    GET  /v1/models        → 获取模型信息
```

Go 后端客户端实现在 `internal/tts/client.go`，封装了 Synthesize / CheckAudioExists / UploadAudio / EnsureAudioUploaded / TestConnection 等方法。

当前不支持其他 TTS 引擎。

---

## 七、数据模型关联总览

```
VsWorkspace
    ├── VsVoiceAsset (全局声音资产)
    └── VsPronunciationDict (全局词典, project_id IS NULL)

VsProject
    ├── VsChapter → VsScriptSegment (脚本片段)
    ├── VsCharacter (角色, 持有 voice_profile_id)
    ├── VsVoiceProfile (声音配置)
    │     └── VsVoiceEmotion (情绪参考音频)
    ├── VsPronunciationDict (项目级词典)
    └── VsGenerationTask (生成任务)
          └── VsAudioClip (音频结果)

VsTTSProvider (TTS 提供商)
    ├── → VsProject.tts_provider_id (项目默认)
    ├── → VsVoiceProfile.tts_provider_id (配置级覆盖)
    ├── → VsVoiceAsset.tts_provider_id (资产级)
    ├── → VsGenerationTask.tts_provider_id (任务记录)
    └── → VsAudioClip.tts_provider_id (生成记录)
```
