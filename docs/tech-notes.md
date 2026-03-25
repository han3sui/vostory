# 技术注意点

本文档记录开发过程中发现的关键技术决策和注意事项，按阶段组织。

---

## 第 1 阶段：主链路

### ID 策略：统一自增

- **所有表**（`vs_*` 业务表 + `sys_*` 系统表）：统一使用 `uint64` 自增主键
- GORM tag：`gorm:"primaryKey;autoIncrement;comment:xxxID"`
- 不再使用 Sonyflake 雪花 ID，service 层 Create 方法中无需手动生成 ID
- `uint64` 自增 ID 在 JS 安全整数范围内（`2^53`），前端无需特殊处理

### 文件导入

- txt 解析使用正则 `^第[零一二三四五六七八九十百千万\d]+[章节回卷集篇]` 匹配章节标题
- docx / epub 格式当前仅支持上传存储，自动解析待后续阶段实现
- 上传路径：`storage/uploads/{projectID}/{timestamp}.{ext}`

---

## 第 2 阶段：可编辑

### 精准填充算法

精准填充（`vs_precise_fill.go`）将 LLM 拆分结果对齐回章节原文：

1. **精确匹配优先**：按片段顺序在原文中 `strings.Index` 查找
2. **模糊匹配兜底**：精确匹配失败时，使用滑动窗口 + LCS（最长公共子序列）算法
3. 匹配阈值：LCS 长度 > 片段长度的 50% 才视为有效匹配
4. 替换后将原 Content 保存到 `OriginalContent` 字段，便于对比回溯

### 发音词典合并规则

`FindEffective` 接口返回项目级 + 全局级合并后的有效词典：

- 项目级词条优先覆盖全局级同名词条
- 全局词典按 `workspace_id` 隔离，`project_id IS NULL` 表示全局

---

## 第 3 阶段：可生成

### LLM 调用协议

**统一使用 OpenAI Chat Completions 协议**（`POST /v1/chat/completions`），不需要为每个厂商写独立适配器。

兼容的厂商包括：OpenAI、Azure OpenAI、DeepSeek、通义千问（阿里）、智谱 GLM、百川、Moonshot（Kimi）、小米 MiMo、零一万物（Yi）、MiniMax、Ollama、vLLM、LM Studio 等。

**关键设计**：

- 用户配置 `api_base_url` 时需填到 `/v1` 级别（如 `https://api.deepseek.com/v1`）
- 厂商差异通过 `custom_params`（JSON）传入，展开到请求参数中（如 `response_format`、`temperature`）
- 项目绑定 `llm_provider_id`，运行时从 provider 取连接参数构造请求

**连通性测试**：

- 不要用 `/v1/models` 端点（很多厂商不支持，会 404）
- 改为发送一次最小的 Chat Completions 请求（`max_tokens: 1`，内容 `"hi"`）
- 自动检测 `api_base_url` 是否已包含 `/v1`，拼接 `/chat/completions`

**非 OpenAI 兼容协议**（Anthropic Messages API、Google Gemini API）：

- MVP 阶段不做原生支持
- 可通过 OpenRouter、OneAPI 等中转网关转成 OpenAI 协议后接入

### LLM 调用层（待实现）

后续开发 LLM 调用服务时，参考 SonicVale 项目的 `LLMEngine` 模式：

```go
// 伪代码示意
type LLMEngine struct {
    apiKey      string
    baseURL     string
    modelName   string
    customParams map[string]interface{}
}

func (e *LLMEngine) ChatCompletion(prompt string) (string, error) {
    // POST {baseURL}/chat/completions
    // Body: { model, messages, ...customParams }
}
```

- 所有业务场景（章节解析、角色抽取、情绪标注等）统一通过此引擎调用
- 每次调用记录到 `vs_llm_log` 表（Prompt 版本、模型、耗时、输入输出摘要）
- 长文分批处理：按可配置字数上限切块，每块独立调用，支持单批次重试

### TTS 调用协议（待实现）

TTS 厂商协议差异较大，不像 LLM 有统一标准，需要设计适配层接口：

- 统一接口：`Synthesize(text, voiceProfile, emotion) -> audioBytes`
- 按 `provider_type` 分发到不同实现
- 情绪 + 强度双维度传参：通过 `VoiceEmotion` 表查找对应情绪的参考音频
