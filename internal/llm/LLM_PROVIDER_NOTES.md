# 🧠 LLM Provider Integration — Developer Notes

> **မြန်မာဘာသာ** — LLM Provider switch system
> Folder: `internal/llm/`

---

## 📋 မာတိကာ

- [ဘာတွေ ပြုလုပ်ထားလဲ](#ဘာတွေ-ပြုလုပ်ထားလဲ)
- [Architecture ပြောင်းလဲပုံ](#architecture-ပြောင်းလဲပုံ)
- [Switch လုပ်နည်း](#switch-လုပ်နည်း)
- [Files ရှင်းလင်းချက်](#files-ရှင်းလင်းချက်)

---

## ✅ ဘာတွေ ပြုလုပ်ထားလဲ

| File | ပြောင်းလဲချက် |
|---|---|
| `internal/llm/llm.go` | **NEW** — LLM interface (contract) |
| `internal/llm/gemini.go` | **NEW** — Google Gemini API client |
| `internal/llm/ollama.go` | **မပြောင်း** — Ollama client (ဟောင်းအတိုင်း) |
| `internal/config/config.go` | `LLMProvider`, `GeminiAPIKey`, `GeminiModel` ထည့် |
| `internal/rag/agent.go` | `*llm.Ollama` → `llm.LLM` interface သို့ ပြောင်း |
| `main.go` | Provider selection logic ထည့် |
| `.env` | `LLM_PROVIDER`, `GEMINI_API_KEY`, `GEMINI_MODEL` ထည့် |

---

## 🏗️ Architecture ပြောင်းလဲပုံ

**ယခင် (fixed):**
```
agent → *Ollama (Ollama တစ်ခုတည်းသာ သုံးနိုင်)
```

**ယခု (flexible):**
```
agent → LLM interface
              ↙              ↘
        *Ollama           *Gemini
      (local PC)       (Google API)
```

### LLM Interface (`llm.go`)

```go
type LLM interface {
    Chat(messages []Message) (string, error)
    ChatStream(messages []Message, onToken func(string) error) error
}
```

> Interface ဆိုတာ **"ဘာတွေ လုပ်နိုင်မလဲ"** ကို သတ်မှတ်တယ် — Ollama ဖြစ်စေ Gemini ဖြစ်စေ ဒီ function ၂ ခု ရှိရမယ်

---

## 🔄 Switch လုပ်နည်း

`.env` ထဲမှာ `LLM_PROVIDER` တစ်ကြောင်းပဲ ပြင်ရတယ်:

### Ollama သုံးရင် (default)
```bash
LLM_PROVIDER=ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=frontiir-ai:latest
```

### Gemini သုံးရင်
```bash
LLM_PROVIDER=gemini
GEMINI_API_KEY=AIzaSyXXXXXXXXXXXXXXXX
GEMINI_MODEL=gemini-2.0-flash-lite
```

> API Key ကို [https://aistudio.google.com](https://aistudio.google.com) မှ ယူပါ

---

### Server Start တဲ့အခါ Log မြင်ရမယ်

```bash
# Ollama သုံးရင်
LLM provider: Ollama (frontiir-ai:latest @ http://localhost:11434)

# Gemini သုံးရင်
LLM provider: Gemini (gemini-2.0-flash-lite)
```

---

## 📁 Files ရှင်းလင်းချက်

### `llm.go` — Interface

```go
type LLM interface {
    Chat(...)       // sync response
    ChatStream(...) // token-by-token stream
}
```

ဒီ interface ကြောင့် `agent.go` သည် Ollama/Gemini ဘယ်ဟာ သုံးနေမှန်း **မသိဘဲ** အလုပ်လုပ်နိုင်တယ်

---

### `ollama.go` — Ollama Client

```
Local PC မှာ run တဲ့ Ollama server နဲ့ communicate
POST http://localhost:11434/api/chat
```

- **ကောင်းတာ:** Free, Private, Internet မလို
- **မကောင်းတာ:** RAM လိုတယ် (2-9 GB)

---

### `gemini.go` — Gemini Client

```
Google Cloud Gemini REST API နဲ့ communicate
POST https://generativelanguage.googleapis.com/v1beta/models/...
```

- **ကောင်းတာ:** RAM မလို, မြန်တယ်, Free tier ရှိ
- **မကောင်းတာ:** Internet လို, Data Google ဆီ ရောက်တယ်

#### Gemini မှာ Format ကွာတဲ့ အပိုင်း

| Ollama | Gemini |
|---|---|
| `"role": "system"` | `system_instruction` (သီးသန့်) |
| `"role": "assistant"` | `"role": "model"` |
| `"role": "user"` | `"role": "user"` |

---

## 💡 Gemini Free Tier Limits

| Limit | တန်ဖိုး |
|---|---|
| Requests per Day | 1,000 RPD |
| Requests per Minute | 15 RPM |
| Tokens per Minute | 250,000 TPM |
| Cost | $0 (Free tier) |

> ⚠️ Free tier မှာ Google က data သုံး improve လုပ်ပိုင်ခွင့် ရတယ် — Customer data ပါတဲ့ production မှာ **မသင့်တော်ဘူး**

---

## 🔮 အနာဂတ်မှာ Provider အသစ် ထပ်ထည့်ချင်ရင်

```
1. internal/llm/openai.go (ဥပမာ) ဖန်တီး
2. LLM interface ကို implement လုပ် (Chat + ChatStream)
3. config.go မှာ fields ထည့်
4. main.go ရဲ့ switch case မှာ ထည့်
5. .env မှာ API key ထည့်
```

Agent code (agent.go) ကို **လုံးဝ မပြင်ရဘဲ** provider အသစ် ထပ်ထည့်နိုင်တယ်

---

*Last updated: 2026-03-04 | Frontiir AI Project*
