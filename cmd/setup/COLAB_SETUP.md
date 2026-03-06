# 🚀 Google Colab မှာ Project Setup လုပ်နည်း

> **ရည်ရွယ်ချက်:** Local PC processing နှေးတဲ့အခါ Google Colab T4 GPU သုံးပြီး Ollama ကို Cloud မှာ run တာ

---

## 📋 မာတိကာ

1. [Zip File ပြင်ဆင်ခြင်း](#step-1--local-pc-မှာ-zip-လုပ်)
2. [Colab Notebook ဖွင့်ခြင်း](#step-2--google-colab-notebook-ဖွင့်ပါ)
3. [Colab Cell တွေ Run ခြင်း](#step-3--colab-cell-တွေမှာ-တစ်ခုချင်း-run-ပါ)
4. [ngrok URL ယူခြင်း](#step-4--output-မှာ-url-မြင်ရမယ်)
5. [Local .env ပြင်ခြင်း](#step-5--local-env-မှာ-ngrok-url-ထည့်ပါ)

---

## Step 1 — Local PC မှာ Zip လုပ်

```bash
# Terminal မှာ run ပါ (AI_CHAT folder ရဲ့ parent directory မှာ)
zip -r AI_CHAT.zip AI_CHAT/ --exclude "AI_CHAT/.env"
```

> `.env` ကို exclude လုပ်ထားတာ — token/password တွေ မပါသွားအောင်

---

## Step 2 — Google Colab Notebook ဖွင့်ပါ

1. [https://colab.research.google.com](https://colab.research.google.com) → **New Notebook**
2. Menu: **Runtime → Change runtime type → T4 GPU** ရွေးပါ

---

## Step 3 — Colab Cell တွေမှာ တစ်ခုချင်း Run ပါ

### Cell 1 — Go Install

```bash
!apt-get install -y golang-go -q
!go version  # verify
```

### Cell 2 — Project Upload

```python
from google.colab import files
uploaded = files.upload()  # AI_CHAT.zip ကို choose လုပ်ပါ
```

### Cell 3 — Unzip

```bash
!unzip AI_CHAT.zip -d /content/
!ls /content/AI_CHAT  # files တွေ မြင်ရမယ်
```

### Cell 4 — .env file ဖန်တီး

```
%%writefile /content/AI_CHAT/.env
LLM_PROVIDER=ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=frontiir-ai:latest
QDRANT_URL=http://localhost:6333
CPE_API_URL=http://localhost:9000/api/cpe
CUSTOMER_API_URL=http://localhost:9000/api/customer
NGROK_TOKEN=YOUR_NGROK_TOKEN_HERE
```

> ⚠️ Real token ကို ဒီ file မှာ မထည့်ပါနှင့် — Colab Cell မှာသာ ထည့်ပါ

### Cell 5 — Setup Script Run

```bash
%cd /content/AI_CHAT
!go run ./cmd/setup --ngrok-token=YOUR_NGROK_TOKEN_HERE --model=frontiir-ai:latest
```

---

## Step 4 — Output မှာ URL မြင်ရမယ်

```
=== Step 8 — Start ngrok tunnel ===
✅ Ollama is public!

  OLLAMA_URL = https://xxxx.ngrok-free.app

  Run your Go app with:
  OLLAMA_URL=https://xxxx.ngrok-free.app go run .
```

---

## Step 5 — Local .env မှာ ngrok URL ထည့်ပါ

```bash
# local PC ရဲ့ .env မှာ ဒီ line ကို ပြင်ပါ
OLLAMA_URL=https://xxxx.ngrok-free.app  # Colab URL ကို ထည့်
```

ပြီးရင် local PC မှာ `go run .` run လိုက်ရင် Colab GPU ကနေ process လုပ်ပေးမယ်!

---

## ⚠️ မှတ်ရန်

| အချက် | အသေးစိတ် |
|---|---|
| Session limit | Colab free tier = **12 နာရီ** ကုန်ရင် session သေတယ် |
| URL ပြောင်း | Session အသစ် start ရင် ngrok URL အသစ် ထွက်တယ် — `.env` ပြန်ပြင်ရမယ် |
| GPU availability | T4 GPU အမြဲ မရနိုင် — busy ဖြစ်ရင် CPU fallback ဖြစ်မယ် |
| Token security | ngrok token ကို GitHub/Colab notebook မှာ **မ commit မလုပ်ပါနှင့်** |

---

*Last updated: 2026-03-04 | Frontiir AI Project*
