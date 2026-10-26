# TTS 合成设计文档

本文档整理 VoStory TTS 合成链路的设计思路，涵盖发音词典、声音配置、基础音色设计、按场景演绎及与 Breeze-TTS-2 的关联。

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
                    │   Breeze-TTS-2 Voice Direction │
                    │  (参考音色 + 演绎指令)          │
                    └──────────┬────────────────────┘
                               │
                        ┌──────▼───────┐
                        │  AudioClip   │
                        │ (生成结果)    │
                        └──────────────┘
```

合成一个脚本片段时，分两条线并行准备：

- **文本侧**：片段原文 → 经发音词典纠音 → 得到可发音文本
- **音频侧**：片段角色 → 声音配置（基础音色）→ 可选的情绪参考音频 → 演绎指令（片段 instruction 或情绪兜底）

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

Breeze-TTS-2 不支持拼音标注，词典的 `phoneme` 字段改为「替换读法」：合成前把原始词直接替换为同音字写法（如"单于"→"蝉于"），再把文本传给 TTS 引擎。

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

## 四、音色与语气：固定基础音色 + 按场景演绎

### 设计思路

角色声音分两层：

| 层 | 存在哪里 | 由谁决定 | 是否变化 |
|----|----------|----------|----------|
| 基础音色 | `VoiceProfile`：参考音频 + 逐字转写 + 种子 | 音色设计（Voice Design）或手动上传 | 整本书固定 |
| 语境语气 | `ScriptSegment.instruction` | 切分时 LLM 结合场景生成，可人工修改 | 每句不同 |

合成时以 Breeze-TTS-2 的 Voice Direction 模式把两层合在一起：参考音频锁定"谁在说"，演绎指令决定"怎么说"。

### 基础音色的生成（Voice Design）

```
角色信息（名称/别名/性别/层级/描述）+ 若干台词
    → LLM（voice_design 模板）→ 音色描述 instruction + 参考台词 sample_text
    → Breeze Voice Design（只传 instruction + text）→ 参考音频
    → 试听满意 → 保存为 VoiceProfile（reference_audio_url / reference_text / tts_params.seed）并绑定角色
```

Voice Design 每次生成的音色都不同，所以只在设计时调用一次，满意后把音频冻结为参考音频；之后所有片段都基于这条参考音频克隆，不再重新设计。

### 演绎指令（instruction）

- 章节切分时 LLM 已有整章上下文，为每个片段输出一句 ≤20 字的演绎描述（语气、语速、音量、停顿）
- 不描述音色、性别，避免指令把音色带偏
- 为空时按情绪类型 + 强度生成兜底指令（见 [tts-integration.md](./tts-integration.md)）；`neutral` / `auto` 不下发指令

### 情绪类型 × 强度的作用

| 维度 | 可选值 | 作用 |
|------|--------|------|
| emotion_type | auto, neutral, happy, sad, angry, fear, surprise, disgust | 演绎指令为空时的兜底；匹配可选的情绪参考音频；筛选统计 |
| emotion_strength | light, medium, strong | 同上 |

### 音色稳定性

- 同一声音配置固定种子：优先 `tts_params.seed`（音色设计时记录），否则用声音配置 ID
- 演绎指令只写语气，不写音色
- 指令遵循强度 `cfg_scale` 默认 4，可在 `tts_params.cfg_scale` 中按声音配置调整

---

## 五、TTS 合成时的查找链路

合成单个脚本片段的完整流程：

```
1. 输入：ScriptSegment
   ├── content          → 原始文本
   ├── character_id     → 角色 ID
   ├── instruction      → 演绎指令
   ├── emotion_type     → 情绪类型
   └── emotion_strength → 情绪强度

2. 文本处理：
   └── 项目发音词典：词 → 替换读法（同音字）

3. 声音查找：
   ├── Character → voice_profile_id → VoiceProfile
   ├── VoiceEmotion WHERE voice_profile_id = X
   │     AND emotion_type = 片段.emotion_type AND emotion_strength = 片段.emotion_strength
   ├── 找到且有参考文本 → 用情绪参考音频
   └── 否则 → 用 VoiceProfile 的基础音色

4. 引擎：Project.tts_provider_id

5. 调用 Breeze-TTS-2：
   ├── text:        纠音后的文本
   ├── ref_audio:   参考音频
   ├── ref_text:    参考音频逐字转写（必填）
   ├── instruction: 片段演绎指令 / 情绪兜底指令 / 空
   ├── cfg_scale:   tts_params.cfg_scale，默认 4
   └── seed:        tts_params.seed，默认声音配置 ID

6. 输出：AudioClip
   ├── audio_url, duration, file_size, format
   ├── voice_profile_id, emotion_type, emotion_strength
   └── tts_provider_id, version, is_current
```

---

## 六、TTS 引擎（Breeze-TTS-2）

语音合成基于 Breeze-TTS-2，需自行部署。Go 后端通过 HTTP 协议调用：

```text
Breeze-TTS-2 HTTP 端点：
    POST /v1/audio/speech  → 合成（multipart：text / ref_audio / ref_text / instruction / cfg_scale / seed）
                             返回 24kHz 单声道 16bit PCM 流
    GET  /v1/model         → 模型信息
```

Go 后端客户端实现在 `internal/tts/client.go`：`Speech` 发起合成并把 PCM 封装为 WAV，`TestConnection` 调用 `/v1/model`。协议细节与部署见 [tts-integration.md](./tts-integration.md)。

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
