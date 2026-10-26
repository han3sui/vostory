# VoStory

`VoStory` 是一个自用的 AI 有声小说 / 广播剧制作工具。

它的目标不是单纯"把文字转成语音"，而是把小说内容结构化成可编辑、可返工的广播剧工程：角色合并去重、为角色自动设计固定的基础音色、按场景语境生成每句台词的语气，最终拼接导出。

产品定位详见 [product-overview.md](./product-overview.md)，技术架构详见 [architecture.md](./architecture.md)，MVP 方案详见 [mvp-plan.md](./mvp-plan.md)。

## 总体架构

```
┌─────────────────────────────────────────────────────┐
│                    Web 前端                          │
│          Vue 3 + TypeScript + Arco Design           │
├─────────────────────────────────────────────────────┤
│                  业务服务层 (Go)                      │
│   Gin + GORM + Wire + Redis + PostgreSQL            │
│   用户权限 │ 项目管理 │ 任务调度 │ 资产管理           │
├──────────────────────┬──────────────────────────────┤
│   AI 服务层           │       存储层                  │
│   LLM Provider       │   PostgreSQL (业务数据)       │
│   Breeze-TTS-2 (TTS) │   Redis (缓存/队列)           │
│                      │   MinIO/S3 (文件/音频)        │
└──────────────────────┴──────────────────────────────┘
```

### 技术栈

- **前端**：Vue 3 + TypeScript + Vite + Arco Design Vue + @easyfe/admin-component
- **后端**：Golang + Gin + GORM + Wire（依赖注入）+ Redis
- **数据库**：PostgreSQL
- **缓存 / 队列**：Redis
- **文件存储**：MinIO 或兼容 S3 的对象存储
- **TTS 服务**：Breeze-TTS-2（自部署，HTTP 流式接口）
- **实时通信**：SSE（Server-Sent Events，任务进度推送）
- **文档**：Swagger（swag）

## 核心领域对象

从第一版开始就需要把核心数据结构抽清楚：

| 对象 | 说明 |
|------|------|
| `Workspace` | 工作空间（项目分组） |
| `Project` | 一本书或一个制作项目 |
| `Chapter` | 章节 |
| `Scene` | 场景 |
| `ScriptSegment` | 脚本片段（旁白、对白、独白），含情绪类型、情绪强度和演绎指令 |
| `Character` | 项目内角色，支持别名、合并（`merged_into_id`）与路人池 |
| `VoiceProfile` | 声音配置：参考音频 + 逐字转写 + 固定种子，即角色的基础音色 |
| `VoiceAsset` | 全局可复用声音资产 |
| `PronunciationDictionary` | 发音词典（项目级 + 全局级） |
| `GenerationTask` | 生成任务，支持子任务/批次追踪 |
| `TaskBatch` | 任务批次，长文分批处理的最小单元 |
| `AudioClip` | 生成后的音频片段 |
| `ExportJob` | 章节导出任务 |
| `LLMProvider` | LLM 服务商配置（API 地址、密钥、模型列表） |
| `TTSProvider` | TTS 服务商配置（API 地址、密钥、能力列表） |
| `PromptTemplate` | Prompt 模板（角色抽取、对白识别、情绪标注等） |
| `LLMLog` | LLM 调用日志审计 |

## MVP 链路

```
导入小说 → 自动切分脚本（含演绎指令）→ 合并角色 → 自动设计音色 → 批量生成 → 人工修正 → 导出章节
```

详细 MVP 范围和开发阶段见 [mvp-plan.md](./mvp-plan.md)。

## 开发阶段规划

### 第 1 阶段：打通主链路

核心目标：项目创建 → 小说导入 → 章节切分 → 脚本片段存储 → 角色抽取存储。

同时完成以下基础设施：

- **数据库表设计**：全部核心领域对象的表结构
- **AI 服务配置模块**：LLMProvider / TTSProvider 的 CRUD 和连通性测试
- **Prompt 模板管理**：内置默认模板 + 用户自定义，按类型分（角色抽取、对白识别、情绪标注等）
- **LLM 调用日志**：记录每次调用的 Prompt 版本、模型、耗时、结果

### 第 2 阶段：可编辑

- 脚本编辑页（片段级编辑、修正说话人、调整类型/情绪/强度/演绎指令）
- 角色编辑页（绑定声音配置、编辑别名、合并角色、归入路人池）
- 发音词典（项目级 + 全局级）
- 精准填充：LLM 拆分台词后自动对齐回原文，确保不丢字不加字

### 第 3 阶段：可生成

- Breeze-TTS-2 对接（自部署，`POST /v1/audio/speech`）
- 音色设计：LLM 撰写音色描述 → Voice Design 生成参考音频 → 保存为角色的基础音色
- 场景演绎：切分时 LLM 为每个片段生成演绎指令，合成时以 Voice Direction 模式传入（为空时按情绪+强度兜底）
- 可选的多情绪参考音频：同一角色特定情绪使用单独的参考音
- 任务队列（Redis，异步生成、重试、超时、状态追踪）
- SSE 实时推送任务进度
- 音频缓存与版本管理
- 单句试听与重生成

### 第 4 阶段：可交付

- 章节音频拼接导出（停顿插入、音量标准化）
- 多轨音频合并（pydub / ffmpeg）
- 导出记录管理
- 失败重试

### 第 5 阶段：可复用

- 全局声音资产库，跨项目复用音色

## 关键设计决策

### 1. LLM 多厂商适配

采用 Provider 策略模式，统一 `LLMProvider` 接口，支持 OpenAI / DeepSeek / Anthropic / Gemini / Ollama / Azure / 阿里云百炼 / 自定义兼容接口。配置存储在数据库，支持运行时切换。

### 2. TTS 引擎（Breeze-TTS-2）

语音合成基于 Breeze-TTS-2，需自行部署。每个角色绑定一条固定的参考音频（基础音色，可由 Voice Design 自动生成），每个片段携带一句演绎指令，合成时以 Voice Direction 模式保持音色、调整语气。详见 [tts-design.md](./tts-design.md)。

### 3. 长文分批处理

长篇小说按可配置的字数上限切块，每块独立调用 LLM，每个批次独立记录状态（`TaskBatch`），支持单批次重试。最后合并多批结果并处理跨批次的角色去重和对话连续性。

### 4. 任务系统

基于 Redis 队列的异步任务系统，支持排队、重试、超时、取消、恢复和状态追踪。通过 SSE（Server-Sent Events）实时推送任务进度到前端，避免轮询。SSE 基于标准 HTTP 长连接，Gin 原生支持，浏览器 `EventSource` API 内置自动重连，部署无需额外配置。

### 5. 资产分层

```
VoiceAsset（全局）→ VoiceProfile（项目内绑定）→ Character（角色使用）
```

角色资产从一开始就做分层设计，预留跨项目复用能力。

### 6. Prompt 模板可配置

角色抽取、对白识别、情绪标注等 Prompt 分开管理，支持版本记录。项目可绑定不同的 Prompt 模板，便于针对不同类型小说调优。

## 最大技术难点

| 难点 | 说明 |
|------|------|
| 角色抽取与去重 | 同一角色可能有多个名字、称呼、代词，误拆会导致声音错配 |
| 台词归属 | 小说里很多对白没有显式说话人，多人连续对话容易分错 |
| 声音一致性 | 同一角色在长篇中必须稳定，跨章节跨项目要能复用 |
| 长篇任务调度 | 分析、生成、导出都需要分批处理、断点续跑和缓存 |
| 局部返工 | 改一句台词不应重跑整章，每个音频片段需可追踪来源和版本 |
| 人工修正效率 | 模型一定会犯错，产品价值取决于"修错有多快" |

## 开发注意点

1. **先设计中间结构**：尽快把原文转成结构化对象，不要反复操作整本原文
2. **模型调用做适配层**：定义统一接口，未来替换模型不推倒重来
3. **从第一天做版本追踪**：脚本片段、角色绑定、发音词典、音频片段都需要变更记录
4. **任务系统独立设计**：生成任务必须异步化，支持队列、重试、超时、中断、恢复
5. **先做高频修正点**：改说话人、改角色映射、改发音、重生成一句、重新导出——把这些做顺
6. **资产分层设计**：项目内角色 / 全局声音资产 / 可复用 voice profile 模板

## 竞品参考

在设计过程中参考了以下同类项目的优秀实践：

| 借鉴点 | 来源 | 应用位置 |
|--------|------|---------|
| LLM 多厂商 Provider 策略映射 | AI-NovelFlow | LLMProvider 设计 |
| Prompt 模板管理系统 | AI-NovelFlow | PromptTemplate 模块 |
| LLM 调用日志审计 | AI-NovelFlow | LLMLog 模块 |
| 情绪 + 强度双维度 | SonicVale | ScriptSegment / VoiceProfile |
| 多情绪参考音频 | SonicVale | VoiceProfile 设计 |
| 实时推送任务进度（原 WebSocket，改用 SSE） | SonicVale | 任务系统 |
| 精准填充 / 文本对齐 | SonicVale | 第 2 阶段脚本编辑 |
| 长文分批处理 + 子任务追踪 | AI-NovelSpeaker-V2 | GenerationTask / TaskBatch |
| 音频多轨合并（pydub/ffmpeg） | AI-NovelFlow / AI-NovelSpeaker-V2 | 第 4 阶段导出 |

### VoStory 的差异化优势

- **服务化架构**：长任务在后台队列执行，TTS 可部署在单独的 GPU 机器上
- **PostgreSQL + Redis**：比 SQLite 更适合并发和生产环境
- **完善的 RBAC**：已有用户/角色/权限/部门/操作日志体系
- **资产分层复用**：VoiceAsset / VoiceProfile 分层，支持跨项目复用
