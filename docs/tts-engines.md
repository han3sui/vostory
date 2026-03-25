# TTS 引擎选型分析

本文档分析可用于 VoStory 的 TTS 引擎方案，评估其与项目需求的匹配度。

VoStory 的核心需求：
- **声音克隆**：每个角色用参考音频定义音色
- **情绪控制**：7 种情绪 × 3 档强度
- **中文为主**：有声小说 / 广播剧场景
- **发音词典**：多音字、专有名词纠音
- **批量生成**：按章节批量合成，非实时对话

---

## 一、引擎总览

| 引擎 | 类型 | 声音克隆 | 情绪控制 | 中文支持 | 成本 | 推荐度 |
|------|------|---------|---------|---------|------|--------|
| **IndexTTS2** | 本地部署 | ✅ 零样本 | ✅ 多模态 | ✅ 原生 | 免费（需 GPU） | ⭐⭐⭐⭐⭐ |
| **GPT-SoVITS** | 本地部署 | ✅ 零样本 | ⚠️ 仅参考音频 | ✅ 原生 | 免费（需 GPU） | ⭐⭐⭐⭐ |
| **Gemini 2.5 TTS** | 在线 API | ❌ 预置音色 | ✅ 自然语言 | ✅ 支持 | 按量付费 | ⭐⭐⭐ |
| **Fish Audio** | 在线 API | ✅ 克隆 | ⚠️ 控制标签 | ✅ 支持 | $15/百万字节 | ⭐⭐⭐⭐ |
| **Azure TTS** | 在线 API | ⚠️ 付费定制 | ✅ SSML | ✅ 优秀 | 按量付费 | ⭐⭐⭐ |
| **Edge TTS** | 免费在线 | ❌ 固定音色 | ❌ 无 | ✅ 支持 | 免费 | ⭐⭐ |
| **OpenAI Edge TTS** | 免费在线 | ❌ 固定音色 | ❌ 无 | ✅ 支持 | 免费 | ⭐⭐ |

---

## 二、详细分析

### 1. IndexTTS2（强烈推荐）

- **仓库**：https://github.com/index-tts/index-tts
- **来源**：B 站（Bilibili）开源
- **类型**：本地部署，零样本语音克隆

**核心能力：**

IndexTTS2 是目前最适合 VoStory 的引擎，原因如下：

1. **零样本声音克隆**：只需一段参考音频即可克隆音色，完美匹配 VoiceProfile 的设计
2. **情绪与音色解耦**：IndexTTS2 的关键突破 —— 可以独立控制音色（timbre prompt）和情绪（style prompt），这正是 VoStory 的 VoiceProfile + VoiceEmotion 架构所需要的
3. **多模态情绪控制**：支持三种情绪输入方式，全部可用于 VoStory：
   - **情绪参考音频**（`emo_audio_prompt`）→ 对应 VoiceEmotion 表的 `reference_audio_url`
   - **情绪向量**（`emo_vector`）→ 8 维浮点数组 `[happy, angry, sad, afraid, disgusted, melancholic, surprised, calm]`
   - **文本情绪描述**（`use_emo_text` / `emo_text`）→ 用自然语言描述情绪
4. **情绪强度控制**（`emo_alpha`）：0.0-1.0 连续值，可映射 light→0.3, medium→0.6, strong→0.9
5. **拼音控制**：支持中文混合拼音标注，可直接对接发音词典

**API 调用方式：**

IndexTTS2 提供 WebUI（Gradio），也可通过 Python 直接调用：

```python
from indextts.infer_v2 import IndexTTS2

tts = IndexTTS2(
    cfg_path="checkpoints/config.yaml",
    model_dir="checkpoints",
    use_fp16=True
)

# 方式 1：参考音频控制情绪（对应 VoiceEmotion 表）
tts.infer(
    spk_audio_prompt='voice_profile_ref.wav',   # 音色参考
    text='酒楼丧尽天良，开始借机竞拍房间',
    output_path='gen.wav',
    emo_audio_prompt='emotion_sad.wav',          # 情绪参考
    emo_alpha=0.6                                # 情绪强度
)

# 方式 2：情绪向量（可由 emotion_type + emotion_strength 映射）
tts.infer(
    spk_audio_prompt='voice_profile_ref.wav',
    text='快躲起来！是他要来了！',
    output_path='gen.wav',
    emo_vector=[0, 0, 0, 0.8, 0, 0, 0, 0],     # fear=0.8
    emo_alpha=0.9
)

# 方式 3：拼音控制（对接发音词典）
text = '之前你做DE5很好'  # DE5 = 得（第五声）
tts.infer(spk_audio_prompt='ref.wav', text=text, output_path='gen.wav')
```

**VoStory 适配层映射：**

```
VoiceProfile.reference_audio_url  → spk_audio_prompt（音色）
VoiceEmotion.reference_audio_url  → emo_audio_prompt（情绪音频）
emotion_strength: light/medium/strong → emo_alpha: 0.3/0.6/0.9
发音词典 phoneme                   → 拼音标注混入文本
```

**部署方式：**

- 本地 GPU 部署（推荐 FP16，降低显存占用）
- 可通过 [Xinference](https://inference.readthedocs.io/en/latest/models/builtin/audio/indextts2.html) 部署为 HTTP API 服务
- 也可自行封装 FastAPI 服务

**硬件要求：** 需要 NVIDIA GPU，推荐 8GB+ 显存（FP16 模式）

---

### 2. GPT-SoVITS

- **仓库**：https://github.com/RVC-Boss/GPT-SoVITS
- **类型**：本地部署，零样本/少样本语音克隆

**核心能力：**

1. **零样本克隆**：提供参考音频 + 参考文本即可克隆
2. **多语言**：中文、英文、日文、韩文、粤语
3. **内置 API 服务**：官方 `api.py` 提供 FastAPI HTTP 接口
4. **流式输出**：支持流式返回音频

**API 调用方式（官方 api.py）：**

```bash
# 启动 API 服务
python api.py

# HTTP 请求
POST /tts
{
    "text": "你好，世界",
    "text_lang": "zh",
    "ref_audio_path": "/path/to/reference.wav",
    "prompt_text": "参考音频对应的文本",
    "prompt_lang": "zh",
    "top_k": 5,
    "top_p": 1,
    "temperature": 1,
    "speed_factor": 1.0
}
# 返回：WAV 音频流
```

**局限性：**

- **无原生情绪控制**：只能通过不同情绪的参考音频间接控制，没有 IndexTTS2 那样的情绪解耦能力
- 情绪效果完全取决于参考音频的质量

**VoStory 适配层映射：**

```
VoiceProfile.reference_audio_url  → ref_audio_path
VoiceProfile.reference_text       → prompt_text
VoiceEmotion.reference_audio_url  → 替换 ref_audio_path（按情绪切换参考音频）
发音词典                           → 文本替换方式（将词替换为注音）
```

---

### 3. Gemini 2.5 TTS

- **文档**：https://ai.google.dev/gemini-api/docs/speech-generation
- **类型**：在线 API（Google Cloud）

**核心能力：**

1. **自然语言情绪控制**：通过 prompt 描述风格、情绪、语速、口音等
2. **多说话人**：单次请求最多 2 个说话人
3. **30 种预置音色**：如 Kore（Firm）、Puck（Upbeat）、Enceladus（Breathy）等
4. **80+ 语言**：包括中文普通话（cmn）

**API 调用方式：**

```bash
curl "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-tts:generateContent" \
  -H "x-goog-api-key: $GEMINI_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [{
      "parts": [{
        "text": "Say angrily: 酒楼丧尽天良，开始借机竞拍房间！"
      }]
    }],
    "generationConfig": {
      "responseModalities": ["AUDIO"],
      "speechConfig": {
        "voiceConfig": {
          "prebuiltVoiceConfig": { "voiceName": "Kore" }
        }
      }
    }
  }'
```

**局限性：**

- **无声音克隆**：只有 30 种预置音色，无法用参考音频定义角色声音
- **情绪控制靠 prompt**：效果不如参考音频精确
- **按量付费**：大量章节合成成本较高
- **网络依赖**：需要访问 Google API

**VoStory 适配层映射：**

```
VoiceProfile → 只能映射到 30 种预置 voiceName 之一
emotion_type + emotion_strength → 转换为自然语言 prompt 前缀
    如 "Say with strong anger:" / "Say with light sadness:"
发音词典 → 拼音标注嵌入文本
```

**适用场景**：适合快速原型验证、旁白/描述类片段（不需要角色音色克隆的场景）。

---

### 4. Fish Audio

- **文档**：https://docs.fish.audio
- **类型**：在线商业 API

**核心能力：**

1. **声音克隆**：支持上传参考音频创建自定义音色
2. **控制标签**：支持情绪和语调控制标签
3. **SDK 支持**：Python SDK（`fish-audio-sdk`）和 JavaScript SDK
4. **流式输出**：WebSocket 实时流式合成

**API 调用方式：**

```python
import fish_audio_sdk

client = fish_audio_sdk.Client(api_key="your_key")

# 使用克隆音色合成
result = client.tts(
    model="s2-pro",
    text="你好，世界",
    voice_id="your_cloned_voice_id",
    format="wav"
)
```

**VoStory 适配层映射：**

```
VoiceProfile → Fish Audio 的 voice_id（需先上传参考音频创建）
emotion_type → 控制标签
发音词典 → 文本替换方式
```

**优势**：商业级质量、简单易用、支持克隆
**劣势**：按量付费（$15/百万字节），长篇小说成本较高

---

### 5. Azure TTS（微软 TTS）

- **文档**：https://learn.microsoft.com/en-us/azure/ai-services/speech-service/rest-text-to-speech
- **类型**：在线商业 API

**核心能力：**

1. **SSML 支持**：完整的 SSML 标记语言，精细控制语速、音量、音调、停顿
2. **中文音色丰富**：多种中文神经网络语音（如 `zh-CN-XiaoxiaoNeural`）
3. **情绪风格**：部分中文音色支持 `<mstts:express-as>` 标签控制情绪
4. **发音词典**：原生支持 `<phoneme>` SSML 标签

**API 调用方式（SSML）：**

```xml
<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis"
       xmlns:mstts="https://www.w3.org/2001/mstts" xml:lang="zh-CN">
  <voice name="zh-CN-XiaoxiaoNeural">
    <mstts:express-as style="angry" styledegree="2">
      酒楼丧尽天良，开始借机竞拍房间！
    </mstts:express-as>
  </voice>
</speak>
```

**VoStory 适配层映射：**

```
VoiceProfile → Azure voice name（固定音色列表，无克隆）
emotion_type → SSML express-as style
emotion_strength → SSML styledegree (0.01-2.0)
    light→0.5, medium→1.0, strong→2.0
发音词典 → SSML <phoneme> 标签（最佳对接方式）
```

**优势**：SSML 生态成熟、发音词典原生支持、情绪风格标签
**劣势**：无声音克隆（Custom Neural Voice 需企业级付费）、音色固定

---

### 6. Edge TTS / OpenAI Edge TTS

- **Edge TTS**：https://github.com/rany2/edge-tts
- **OpenAI Edge TTS**：https://github.com/travisvn/openai-edge-tts
- **类型**：免费在线（使用微软 Edge 浏览器的 TTS 服务）

**核心能力：**

Edge TTS 是微软 Edge 浏览器内置的在线 TTS 服务的 Python 封装，OpenAI Edge TTS 则进一步将其包装为 OpenAI 兼容的 `/v1/audio/speech` 接口。

1. **完全免费**：无需 API Key，无调用限制
2. **多语言多音色**：支持大量语言和音色（如 `zh-CN-XiaoxiaoNeural`）
3. **语速/音量/音调调节**：通过 `--rate`、`--volume`、`--pitch` 参数

**Edge TTS 使用：**

```python
import edge_tts
import asyncio

async def synthesize():
    communicate = edge_tts.Communicate(
        text="你好，世界",
        voice="zh-CN-XiaoxiaoNeural",
        rate="+0%",
        volume="+0%",
        pitch="+0Hz"
    )
    await communicate.save("output.mp3")

asyncio.run(synthesize())
```

**OpenAI Edge TTS 使用（Docker 部署后）：**

```bash
docker run -d -p 5050:5050 travisvn/openai-edge-tts:latest

curl -X POST http://localhost:5050/v1/audio/speech \
  -H "Content-Type: application/json" \
  -d '{
    "input": "你好，世界",
    "voice": "zh-CN-XiaoxiaoNeural",
    "response_format": "mp3",
    "speed": 1.0
  }' --output speech.mp3
```

**局限性：**

- **无声音克隆**：只能用微软预置音色
- **无情绪控制**：不支持 SSML 的 `express-as`（微软限制了自定义 SSML）
- **音质一般**：相比 Azure 正式 API 质量略低
- **稳定性风险**：非官方 API，微软可能随时变更

**VoStory 适配层映射：**

```
VoiceProfile → 微软预置 voice name
emotion_type → 无法映射（不支持）
emotion_strength → 无法映射
发音词典 → 文本替换方式（不支持 SSML phoneme）
```

**适用场景**：开发调试、Demo 演示、对音质和情绪要求不高的旁白/描述片段。

---

## 三、推荐方案

### MVP 阶段推荐组合

```
┌─────────────────────────────────────────────────┐
│              VoStory TTS 适配层                   │
├──────────┬──────────┬──────────┬────────────────┤
│ IndexTTS2│GPT-SoVITS│ Edge TTS │  Fish Audio    │
│ (主力)   │ (备选)    │ (调试)   │  (商业备选)    │
│ 本地GPU  │ 本地GPU   │ 免费     │  在线付费      │
└──────────┴──────────┴──────────┴────────────────┘
```

| 优先级 | 引擎 | 角色 | 理由 |
|--------|------|------|------|
| **P0** | **IndexTTS2** | 主力引擎 | 声音克隆 + 情绪解耦 + 拼音控制，与 VoStory 架构完美匹配 |
| **P1** | **GPT-SoVITS** | 备选引擎 | 成熟稳定，社区活跃，API 现成，作为 IndexTTS2 的替代方案 |
| **P2** | **Edge TTS** | 开发调试 | 免费无门槛，用于开发阶段快速验证合成流程 |
| **P3** | **Fish Audio** | 商业备选 | 无 GPU 用户的云端方案，质量好但有成本 |

### 不推荐用于核心功能的

| 引擎 | 原因 |
|------|------|
| Gemini 2.5 TTS | 无声音克隆，30 种预置音色无法满足角色定制需求 |
| Azure TTS | 声音克隆需企业级付费，普通用户只能用固定音色 |
| OpenAI Edge TTS | 与 Edge TTS 本质相同，只是多了 OpenAI 兼容接口包装 |

---

## 四、适配层实现建议

### IndexTTS2 适配器

IndexTTS2 没有官方 HTTP API，需要自行封装。建议方案：

**方案 A：Xinference 部署（推荐）**

Xinference 已内置 IndexTTS2 支持，一行命令部署为 HTTP 服务：

```bash
xinference launch --model-name IndexTTS2 --model-type audio
```

**方案 B：自建 FastAPI 服务**

在 IndexTTS2 仓库基础上封装一个轻量 HTTP API：

```python
# 伪代码 — IndexTTS2 HTTP 服务
@app.post("/v1/tts")
async def synthesize(req: TTSRequest):
    tts.infer(
        spk_audio_prompt=req.reference_audio,
        text=req.text,
        output_path=output_file,
        emo_audio_prompt=req.emotion_audio,     # 可选
        emo_vector=req.emotion_vector,           # 可选
        emo_alpha=req.emotion_alpha,             # 可选
    )
    return FileResponse(output_file)
```

### VoStory 后端适配器接口

```go
type TTSAdapter interface {
    Synthesize(ctx context.Context, req SynthesizeRequest) (*SynthesizeResult, error)
    TestConnection(ctx context.Context) error
}

type SynthesizeRequest struct {
    Text              string            // 纠音后的文本
    ReferenceAudio    string            // 音色参考音频（VoiceProfile）
    ReferenceText     string            // 参考音频文本
    EmotionAudio      string            // 情绪参考音频（VoiceEmotion，可选）
    EmotionType       string            // 情绪类型
    EmotionStrength   string            // 情绪强度
    Params            map[string]any    // 引擎特定参数
}
```

### 情绪映射表（IndexTTS2）

VoStory 的 7 种情绪映射到 IndexTTS2 的 8 维向量：

```go
// [happy, angry, sad, afraid, disgusted, melancholic, surprised, calm]
var emotionVectorMap = map[string][8]float64{
    "neutral":  {0, 0, 0, 0, 0, 0, 0, 1.0},
    "happy":    {1.0, 0, 0, 0, 0, 0, 0, 0},
    "sad":      {0, 0, 1.0, 0, 0, 0, 0, 0},
    "angry":    {0, 1.0, 0, 0, 0, 0, 0, 0},
    "fear":     {0, 0, 0, 1.0, 0, 0, 0, 0},
    "surprise": {0, 0, 0, 0, 0, 0, 1.0, 0},
    "disgust":  {0, 0, 0, 0, 1.0, 0, 0, 0},
}

var strengthAlphaMap = map[string]float64{
    "light":  0.3,
    "medium": 0.6,
    "strong": 0.9,
}
```

### 发音词典对接

| 引擎 | 对接方式 |
|------|---------|
| IndexTTS2 | 将 `phoneme` 转为拼音标注混入文本（如 `重DE5庆` → `CHONG2QING4`） |
| GPT-SoVITS | 文本替换：将词替换为注音文本 |
| Edge TTS | 文本替换（不支持 SSML phoneme） |
| Azure TTS | SSML `<phoneme>` 标签（最佳方案） |
| Fish Audio | 文本替换 |

---

## 五、开发优先级

1. **先用 Edge TTS 跑通合成流程**：免费无门槛，验证任务队列、音频存储、进度推送等基础设施
2. **接入 IndexTTS2**：部署 Xinference 或自建 API，实现声音克隆 + 情绪控制的完整能力
3. **按需接入其他引擎**：根据用户反馈和部署环境，逐步添加 GPT-SoVITS、Fish Audio 等适配器
