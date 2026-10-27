# VoStory 技术架构

## 技术栈

- 前端：`Vue 3 + TypeScript + Vite + Arco Design Vue`
- 服务端：`Golang (Gin + GORM + Wire)`
- TTS 服务：`Breeze-TTS-2`（自部署，HTTP 流式接口）
- 数据库：`PostgreSQL`
- 缓存 / 队列：`Redis`
- 对象存储：`MinIO` 或兼容 S3 的对象存储
- 实时通信：`SSE`（Server-Sent Events）

## 总体架构

```
┌─────────────────────────────────────────────────────┐
│                    Web 前端                          │
│          Vue 3 + TypeScript + Arco Design           │
│                                                     │
│  项目管理 │ 脚本编辑 │ 角色管理 │ 任务监控 │ 导出    │
├───────────────────────┬─────────────────────────────┤
│      REST API         │          SSE                │
│   (业务 CRUD)         │   (任务进度实时推送)         │
├───────────────────────┴─────────────────────────────┤
│                 业务服务层 (Go)                       │
│                                                     │
│  认证权限 │ 项目章节 │ 脚本片段 │ 角色资产            │
│  任务调度 │ 文件管理 │ AI配置  │ 审计日志            │
├──────────────────────┬──────────────────────────────┤
│   AI 服务层           │       存储层                 │
│                      │                              │
│  LLM 适配层          │   PostgreSQL (业务数据)       │
│    ├ OpenAI          │   Redis (缓存/任务队列)       │
│    ├ DeepSeek        │   MinIO/S3 (音频/文件)        │
│    ├ Anthropic       │                              │
│    ├ Gemini          │                              │
│    ├ Ollama          │                              │
│    ├ Azure           │                              │
│    ├ 阿里云百炼       │                              │
│    └ 自定义兼容接口   │                              │
│                      │                              │
│  TTS 引擎            │                              │
│    └ Breeze-TTS-2    │                              │
│      (GPU 自部署)     │                              │
└──────────────────────┴──────────────────────────────┘
```

### 1. Web 前端

负责：

- 项目列表与项目详情
- 章节与脚本编辑
- 角色与声音管理
- AI 服务配置（LLM / TTS 提供商管理）
- Prompt 模板管理
- 生成任务监控（SSE 实时进度）
- 音频试听和导出管理

### 2. 业务服务层

由 Go 实现，负责：

- 登录认证与权限控制（RBAC）
- 项目与章节管理
- 脚本片段管理
- 角色资产管理（项目级 + 全局级）
- AI 服务配置管理（LLMProvider / TTSProvider）
- Prompt 模板管理
- 任务调度与生命周期管理（Redis 队列）
- SSE 连接管理与事件推送
- 文件元数据管理
- LLM 调用日志审计
- 操作审计日志

### 3. AI 服务层

负责：

- **LLM 适配层**：统一接口，策略模式映射多厂商 Provider（Go 实现）
- **文本分析**：章节切分、段落识别
- **人物抽取**：角色识别、别名归并
- **台词归属**：说话人判定
- **演绎指令**：切分时按场景为每个片段生成语气/语速描述
- **TTS 引擎**：基于 Breeze-TTS-2（自部署），Go 后端通过 HTTP 协议调用；音色设计（Voice Design）与按场景演绎（Voice Direction）
- **精准填充**：LLM 输出对齐回原文

### 4. 存储层

负责：

- 业务关系数据（PostgreSQL）
- 项目脚本数据（PostgreSQL）
- 声音资产元数据（PostgreSQL）
- 任务队列与缓存（Redis）
- 生成的音频片段（MinIO / S3）
- 最终导出文件（MinIO / S3）

## 核心领域对象

### 业务对象

| 对象 | 说明 |
|------|------|
| `Workspace` | 工作空间（项目分组） |
| `Project` | 一本书或一个制作项目 |
| `Chapter` | 章节 |
| `Scene` | 场景 |
| `ScriptSegment` | 脚本片段（旁白/对白/独白），含演绎指令 |
| `Character` | 项目内角色，支持别名、层级（主角/配角/路人） |
| `VoiceProfile` | 声音配置，绑定角色与音色 |
| `VoiceAsset` | 全局可复用声音资产 |
| `PronunciationDictionary` | 发音词典（项目级 + 全局级） |
| `AudioClip` | 生成后的音频片段 |
| `ExportJob` | 章节导出任务 |

### 任务与调度

| 对象 | 说明 |
|------|------|
| `GenerationTask` | 生成任务（文本解析、TTS 生成等） |
| `TaskBatch` | 任务批次，长文分批处理的最小单元，可独立重试 |

### AI 配置

| 对象 | 说明 |
|------|------|
| `LLMProvider` | LLM 服务商配置（API 地址、密钥、模型列表、自定义参数） |
| `TTSProvider` | TTS 服务商配置（API 地址、密钥、能力列表） |
| `PromptTemplate` | Prompt 模板（类型：character_extract / dialogue_parse / voice_design 等），支持版本 |
| `LLMLog` | LLM 调用日志（Prompt 版本、模型、耗时、输入输出摘要） |

## 关键技术设计

### LLM 多厂商适配

采用 Provider 策略模式，Go 侧定义统一接口，Python AI 服务实现具体适配器：

```text
LLMProvider interface {
    ChatCompletion(prompt, model, params) -> structured_result
    TestConnection() -> bool
}
```

支持的 Provider 类型：
- `openai`：OpenAI 官方
- `deepseek`：DeepSeek（OpenAI 兼容）
- `anthropic`：Anthropic Claude
- `gemini`：Google Gemini
- `ollama`：本地 Ollama
- `azure`：Azure OpenAI（OpenAI 兼容）
- `aliyun`：阿里云百炼（OpenAI 兼容）
- `custom`：用户自定义 OpenAI 兼容接口

配置存储在数据库（`LLMProvider` 表），支持运行时切换，无需重启服务。

### TTS 引擎（Breeze-TTS-2）

语音合成基于 Breeze-TTS-2，需自行部署。Go 后端通过 HTTP 协议调用：

```text
Breeze-TTS-2 HTTP 端点：
    POST /v1/audio/speech  → 合成语音（multipart：text / ref_audio / ref_text / instruction / cfg_scale / seed）
                             返回 24kHz 单声道 16bit PCM 流，后端封装为 WAV
    GET  /v1/model         → 模型信息（用于连通性测试）
```

音色与语气设计：
- `VoiceProfile` 保存角色的基础音色：参考音频 + 逐字转写 + 固定种子（`tts_params.seed`）
- 基础音色可由 Voice Design 自动生成：LLM 根据角色信息写音色描述 → 只传 `instruction` 生成参考音频
- `ScriptSegment.instruction` 是该句的演绎指令，由切分时 LLM 结合场景生成，可人工修改
- 合成时以 Voice Direction 模式传入参考音频 + 演绎指令；指令为空时只传参考音频，按文本自然演绎

详见 [tts-design.md](./tts-design.md)。

### 长文分批处理

```text
原文 → split_by_max_chars → [batch_1, batch_2, ..., batch_n]
                                  ↓
                          每批独立调用 LLM
                                  ↓
                          TaskBatch 记录状态
                                  ↓
                      merge_batch_outputs → 合并结果
                                  ↓
                      跨批次角色去重 + 对话连续性处理
```

- 按可配置字数上限切块（参考标点和段落边界）
- 每个 `TaskBatch` 独立记录输入文本、字数、状态、输出结果
- 单批次失败可独立重试，不影响其他批次
- 合并阶段处理跨批次的角色名称统一和对话归属连续性

### 任务系统

基于 Redis 队列：

```text
API 请求 → 创建 GenerationTask → 入 Redis 队列
                                      ↓
                              Task Worker 消费
                                      ↓
                              执行（LLM / TTS / 后处理）
                                      ↓
                              更新 DB 状态
                                      ↓
                              SSE 推送进度
```

任务状态机：`pending → running → completed / failed`

支持：排队、重试（可配置次数）、超时、取消、恢复、状态追踪。

SSE 推送事件类型（前端通过 `EventSource` API 监听）：
- `task_progress`：进度百分比更新
- `task_completed`：任务完成
- `task_failed`：任务失败
- `batch_completed`：单批次完成

SSE 优势：基于标准 HTTP 长连接，Gin 原生支持（`c.Stream` + `c.SSEvent`），浏览器 `EventSource` 内置自动重连和 `Last-Event-ID`，部署时无需 Nginx 额外配置 WebSocket upgrade。

### 资产分层

```text
VoiceAsset（全局声音资产库）
    └── VoiceProfile（项目内声音配置）
            ├── 默认参考音频
            └── Character（角色绑定）
```

- `VoiceAsset`：全局级，跨项目复用
- `VoiceProfile`：项目级，绑定具体角色，可引用全局 VoiceAsset
- 参考音频支持上传、试听、替换、版本记录

### Prompt 模板管理

```text
PromptTemplate
    ├── type: character_extract    # 角色抽取
    ├── type: dialogue_parse       # 对白识别
    ├── type: scene_split          # 场景切分
    └── type: text_correct         # 文本精准填充
```

- 内置默认模板，启动时同步到数据库
- 支持用户自定义和调优
- 项目可绑定不同的 Prompt 模板
- 记录每次 LLM 调用使用的 Prompt 版本（写入 `LLMLog`）

## 关键技术难点

### 人物抽取与去重

- 同一角色会有多个名字、称呼和代词
- 一旦角色拆错或合错，后面的声音绑定都会出问题

### 台词归属

- 小说里很多对白没有显式说话人
- 多人连续对话场景最容易误判

### 声音一致性

- 同一角色必须在长篇内容中保持稳定
- 声音资产需要支持跨项目复用

### 长篇任务处理

- 长篇内容必须分块处理
- 需要队列、重试、恢复和状态追踪
- 处理过程必须异步化，不能绑在单个请求上
- 跨批次的角色去重和对话连续性需要额外处理

### 局部返工

- 改一句话不能要求整章重跑
- 音频片段需要有清晰的来源和版本关系

### 人工修正效率

- 模型一定会犯错
- 产品价值很大程度上取决于"修错是否足够快"

## 模型策略

### LLM 策略

推荐采用"本地模型 + 在线模型"的混合方案：

- **本地模型**（Ollama 等）适合：人物抽取、对白分类、别名归并、高频局部修正。成本可控、响应快、数据不出境。
- **在线模型**（OpenAI / DeepSeek 等）适合：复杂语义判断、最终成品精修。效果更强、可快速接入。
- 通过 LLMProvider 配置灵活切换，无需改代码。

### TTS 策略

语音合成统一使用 Breeze-TTS-2，需自行部署。选择它的原因：

- 一个模型覆盖三种模式：音色克隆、音色设计（文字描述生成新音色）、音色导演（保持音色、用自然语言调整语气）
- "固定基础音色 + 每句演绎指令"正好对应「角色声音一致、语气随场景变化」的需求
- 中英双语，支持 `[笑]`、`[叹气]` 等中文声音事件标签
- 普通推理约 8GB 显存，单卡即可部署

注意：Breeze-TTS-2 的开放权重与自部署产出仅限研究与非商业用途，本项目仅供个人使用。

## 开发注意点

### 先构建结构化中间层

不要反复直接处理整本原文，应该尽快转成结构化对象：章节 → 场景 → 脚本片段 → 角色 → 声音映射 → 音频片段。

### 模型调用必须可替换

统一接口定义：

```text
AnalyzeText(input) -> structured_result
GenerateSpeech(segment, voice_profile, instruction) -> audio_clip
PostProcessAudio(clips) -> final_audio
```

### 从第一天开始做版本追踪

至少要能追踪这些对象的变化：脚本片段、说话人绑定、发音词典、音频片段、导出结果。

### 任务系统要独立设计

任务系统需要支持：排队、重试、超时、取消、恢复、状态跟踪、SSE 实时推送。

### 区分全局资产和项目内资产

预留分层：项目内角色 → 全局可复用声音资产 → 可复用的声音配置模板。

### 使用边界

本项目仅供个人使用：导入的小说与生成的音频不对外分发；Breeze-TTS-2 模型权重与产出受其非商业许可约束。
