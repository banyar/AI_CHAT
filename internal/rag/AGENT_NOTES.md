# 🤖 Frontiir AI Agent — Developer Notes
basePrompt ဆိုတာ ဘာလဲ၊ ဘယ်လိုအလုပ်လုပ်လဲ
ဘာလဲ? (What)
basePrompt ဆိုတာ LLM ကို စတင်မှာကြားတဲ့ "သင်ကြားပေးချက်" ဖြစ်တယ်။

Human မှာ employee ကို "သင်တစ်ဦးသည် Frontiir customer support ဖြစ်သည်..." လို့ briefing ပေးသလိုပဲ၊ LLM ကိုလည်း chat စတင်တိုင်း ဒီ instruction ကို ပေးတယ်။

ဘယ်အချိန် အလုပ်လုပ်လဲ? (When)
Chat တိုင်း — user message တစ်ကြိမ် ပေးပို့တဲ့အချိန်မှာ:


User message တစ်ခု ရောက်လာ
          ↓
basePrompt          ← အမြဲပါတယ် (ပထမဆုံး)
    +
RAG context         ← ဆင်တူ docs တွေ ထပ်ထည့်
    +
Memory history      ← ယခင် conversation
    +
User message        ← ယခု မေးခွန်း
          ↓
LLM ထံ ပေးပို့
agent.go:69-75 မှာ ဒီလို build လုပ်တယ်:


systemPrompt := basePrompt          // အမြဲ basePrompt နဲ့ စတယ်

if ctx := buildContext(...); ctx != "" {
    systemPrompt = basePrompt + "\n\nContext:\n" + ctx  // RAG/CPE/Customer data ထပ်တယ်
}

messages := []Message{
    {Role: "system",    Content: systemPrompt},   // ← basePrompt ဒီမှာ
    ...memory history...,
    {Role: "user",      Content: userMsg},
}
basePrompt ရဲ့ အပိုင်းတစ်ခုချင်း ရှင်းလင်းချက်
အပိုင်း 1 — Role သတ်မှတ်ချက်

You are Frontiir AI Assistant — a helpful customer support
assistant for Frontiir, Myanmar's leading fiber internet provider.
ဘာကြောင့်: LLM ကို "ဘယ်သူလဲ" သတ်မှတ်ပေးတယ်။
မသတ်မှတ်ရင် LLM က general assistant ဖြစ်နေပြီး Frontiir topic မဟုတ်တာတွေ ဖြေနိုင်တယ်။

အပိုင်း 2 — ဘာသာစကား Rule

- If the user writes in Burmese (မြန်မာ), reply ONLY in pure Burmese.
- If the user writes in English, reply ONLY in English.
ဘာကြောင့်: LLM တွေသည် automatically ဘာသာပေါင်းစပ်တတ်တယ်။

မရှိရင် ဖြစ်နိုင်တဲ့ ပြဿနာ:


User:  "internet မကောင်းဘူး"
AI:    "Your internet might be experiencing issues. 
        ကျေးဇူးပြု၍ router ကို restart please do it."  ← ရောနေတယ်
ရှိရင်:


User:  "internet မကောင်းဘူး"
AI:    "Router ကို restart လုပ်ပြီး ၅ မိနစ် စောင့်ပါ။"  ← သန့်တယ်
အပိုင်း 3 — Mixed Word တားမြစ်ချက်

- Never combine two languages in one word.
  Write "ဥပမာ" OR "example" — never "ဥပမable".
ဘာကြောင့်: LLM တွေသည် တစ်ခါတရံ word ကိုယ်တိုင် ရောနေတဲ့ artifact ထုတ်တယ်။


မရှိရင်:  "ဥပမable တစ်ခုပြောရရင်..."   ← ဖြစ်နိုင်တယ်
ရှိရင်:   "ဥပမာ တစ်ခုပြောရရင်..."      ← မဖြစ်တော့ဘူး
ဒါကြောင့် postprocess/cleaner.go လည်း ထပ်ထည့်ထားတယ် — prompt ရဲ့ second line of defense အနေနဲ့

အပိုင်း 4 — Technical Terms ခွင့်ပြုချက်

- Technical terms (CPE, ID, router, ONT, signal, status)
  may stay in English inside a Burmese reply.
ဘာကြောင့်: "CPE", "router", "ONT" ဆိုတာတွေကို မြန်မာလို ဘာသာပြန်ရင် ဆိုးလာမယ်။


မရှိရင်:  "သင့် ဖောက်သည် နေရာ ပစ္စည်း မကောင်းဘူး"  ← CPE ကို ဘာသာပြန်ထားတာ
ရှိရင်:   "သင့် CPE device မကောင်းဘူး"              ← ရှင်းတယ်
အပိုင်း 5 — Example Format

- Use "ဥပမာ -" (not "ဥပမာ:" or "example:") when giving examples
ဘာကြောင့်: Output format ကို consistent ထားဖို့ — response တွေ ညီတူ ပုံစံ ဖြစ်အောင်

Summary — basePrompt ရဲ့ Role

basePrompt = AI ရဲ့ "Job Description" + "လုပ်ပိုင်ခွင့် စည်းကမ်း"
                         ↓
          Chat တိုင်း System Message အနေနဲ့ LLM ကို ပေးတယ်
                         ↓
     LLM က "Frontiir Customer Support" role ကို ဆောင်ပြီး ဖြေတယ်
မရှိရင်	ရှိရင်
ဘာ topic မဆို ဖြေတယ်	Frontiir topic ပဲ ဖြေတယ်
ဘာသာ ရောနေတယ်	ဘာသာ သန့်တယ်
Mixed words ဖြစ်တယ်	Pure Burmese/English
Format မညီ	Format ညီ
if qdrant database has no data when user ask some chat about frontiir info at the time how llm will work just answer olny don't change any code and explain to burmese
Qdrant မှာ data မရှိရင် LLM ဘယ်လိုအလုပ်လုပ်လဲ
ဖြစ်ပျက်တဲ့ flow

User: "Frontiir package တွေ ဘာတွေ ရှိလဲ?"
          ↓
Embedder: message → vector
          ↓
Qdrant Search → docs[] = [] (empty — data မရှိ)
          ↓
buildContext([], userMsg)
          ↓
context = "" (empty string)
          ↓
if ctx != ""  → FALSE ဖြစ်တယ်
          ↓
systemPrompt = basePrompt တစ်ခုတည်း
(RAG context မပါဘဲ)
          ↓
LLM ထံ ပေးပို့
agent.go:70-75 မှာ ဒီ logic ရှိတယ်:


systemPrompt := basePrompt  // default

if ctx := buildContext(docs, userMsg); ctx != "" {
    // Qdrant data ရှိမှပဲ ဒီထဲ ဝင်တယ်
    systemPrompt = basePrompt + "\n\nContext:\n" + ctx
}
// data မရှိရင် → basePrompt တစ်ခုတည်းနဲ့ LLM ကို ခေါ်တယ်
LLM ဘာပြုလုပ်မလဲ
Qdrant data မရှိရင် LLM သည် သူ့ကိုယ်ပိုင် pre-trained knowledge ကိုသာ သုံးပြီး ဖြေတယ်။


basePrompt:  "You are Frontiir AI Assistant..."
                      ↓
LLM ထဲမှာ သိပြီးသား Frontiir နဲ့ ဆင်တူတဲ့ Myanmar ISP knowledge
+ general internet troubleshooting knowledge
                      ↓
ဖြေဆိုတယ် — accuracy နည်းနိုင်သော်လည်း ဖြေနိုင်သေးတယ်
ဥပမာ နှိုင်းယှဉ်ချက်
အခြေအနေ	LLM ဖြေပုံ
Qdrant data ရှိ	"Frontiir Home Package သည် 5,000 ကျပ်မှ..." (တိကျတယ်)
Qdrant data မရှိ	"Frontiir ၏ package များသည် ယေဘုယျအားဖြင့်..." (မတိကျနိုင်)
အရေးကြီးတဲ့ နားလည်မှု
Qdrant data သည် LLM ကို ပိုတိကျစေဖို့ ထည့်ပေးတာသာ ဖြစ်တယ် — LLM ကို run ဖို့ မလိုအပ်ဘူး။


Qdrant data = Extra study notes
LLM         = Student ကိုယ်တိုင်

notes ရှိ  → notes ကြည့်ပြီး ဖြေ (ပိုတိကျ)
notes မရှိ → ကိုယ်သိတာနဲ့ ဖြေ (ကိုယ်တိုင် knowledge)
ဒါကြောင့် production မှာ training data (JSONL) တွေကို ingest လုပ်ထားဖို့ အရေးကြီးတယ် — LLM ရဲ့ ဖြေကြားမှု accuracy သည် Qdrant ထဲ data ပေါ်မှာ တိုက်ရိုက် မူတည်တယ်။
> **မြန်မာဘာသာ** — Developer Reference Guide
> File: `internal/rag/agent.go`

---

## 📋 မာတိကာ

- [basePrompt ဆိုတာ ဘာလဲ](#baseprompt-ဆိုတာ-ဘာလဲ)
- [Qdrant data မရှိရင် ဘာဖြစ်မလဲ](#qdrant-data-မရှိရင်-ဘာဖြစ်မလဲ)
- [Message Build လုပ်ပုံ](#message-build-လုပ်ပုံ)

---

## 🟡 basePrompt ဆိုတာ ဘာလဲ

> **LLM ကို စတင်မှာကြားတဲ့ "သင်ကြားပေးချက်"** ဖြစ်တယ်။
> Human မှာ employee ကို briefing ပေးသလိုပဲ — chat တိုင်း ပထမဆုံး ပေးပို့တယ်။

---

### ⏱️ ဘယ်အချိန် အလုပ်လုပ်လဲ

```
User message တစ်ခု ရောက်လာ
        ↓
basePrompt          ← အမြဲပါတယ် (ပထမဆုံး)
    +
RAG context         ← ဆင်တူ docs တွေ ထပ်ထည့်
    +
Memory history      ← ယခင် conversation
    +
User message        ← ယခု မေးခွန်း
        ↓
LLM ထံ ပေးပို့
```

`agent.go` မှာ build လုပ်ပုံ:

```go
systemPrompt := basePrompt          // အမြဲ basePrompt နဲ့ စတယ်

if ctx := buildContext(...); ctx != "" {
    systemPrompt = basePrompt + "\n\nContext:\n" + ctx
}

messages := []Message{
    {Role: "system",  Content: systemPrompt}, // ← basePrompt ဒီမှာ
    // ...memory history...
    {Role: "user",    Content: userMsg},
}
```

---

### 🔵 အပိုင်း ၁ — Role သတ်မှတ်ချက်

```
You are Frontiir AI Assistant — a helpful customer support
assistant for Frontiir, Myanmar's leading fiber internet provider.
```

| | မသတ်မှတ်ရင် | သတ်မှတ်ရင် |
|---|---|---|
| AI role | General assistant | Frontiir Support |
| Topic | ဘာမဆို ဖြေတယ် | Frontiir topic ပဲ ဖြေတယ် |

---

### 🟢 အပိုင်း ၂ — ဘာသာစကား Rule

```
- If the user writes in Burmese (မြန်မာ), reply ONLY in pure Burmese.
- If the user writes in English, reply ONLY in English.
```

**Rule မရှိရင်** ❌
```
User: "internet မကောင်းဘူး"
AI:   "Your internet might be issues. ကျေးဇူးပြု၍ restart please do it."
```

**Rule ရှိရင်** ✅
```
User: "internet မကောင်းဘူး"
AI:   "Router ကို restart လုပ်ပြီး ၅ မိနစ် စောင့်ပါ။"
```

---

### 🟠 အပိုင်း ၃ — Mixed Word တားမြစ်ချက်

```
- Never combine two languages in one word.
  Write "ဥပမာ" OR "example" — never "ဥပမable".
```

| Rule မရှိရင် ❌ | Rule ရှိရင် ✅ |
|---|---|
| `"ဥပမable တစ်ခုပြောရရင်..."` | `"ဥပမာ တစ်ခုပြောရရင်..."` |

> ⚠️ **Double Protection:** `postprocess/cleaner.go` သည် prompt fail ဖြစ်ရင် second line of defense အနေနဲ့ ထပ်စစ်တယ်

---

### 🔴 အပိုင်း ၄ — Technical Terms ခွင့်ပြုချက်

```
- Technical terms (CPE, ID, router, ONT, signal, status)
  may stay in English inside a Burmese reply.
```

| Rule မရှိရင် ❌ | Rule ရှိရင် ✅ |
|---|---|
| `"သင့် ဖောက်သည် နေရာ ပစ္စည်း မကောင်းဘူး"` | `"သင့် CPE device မကောင်းဘူး"` |

---

### 🟣 အပိုင်း ၅ — Example Format

```
- Use "ဥပမာ -" (not "ဥပမာ:" or "example:") when giving examples
```

Output format ကို **consistent** ထားဖို့ — response တွေ ညီတူ ပုံစံ ဖြစ်အောင်

---

### 📊 basePrompt Summary

```
basePrompt = AI ရဲ့ "Job Description" + "လုပ်ပိုင်ခွင့် စည်းကမ်း"
                          ↓
         Chat တိုင်း System Message အနေနဲ့ LLM ကို ပေးတယ်
                          ↓
    LLM က "Frontiir Customer Support" role ဆောင်ပြီး ဖြေတယ်
```

| အချက် | မရှိရင် | ရှိရင် |
|---|---|---|
| Topic | ဘာ topic မဆို ဖြေတယ် | Frontiir topic ပဲ |
| ဘာသာ | ရောနေတယ် | သန့်တယ် |
| Mixed words | ဖြစ်နိုင်တယ် | မဖြစ်တော့ |
| Format | မညီ | ညီ |

---

## 🗄️ Qdrant data မရှိရင် ဘာဖြစ်မလဲ

### Flow

```
User: "Frontiir package တွေ ဘာတွေ ရှိလဲ?"
          ↓
Embedder: message → vector
          ↓
Qdrant Search → docs[] = []  ← empty (data မရှိ)
          ↓
buildContext([], userMsg) → "" (empty string)
          ↓
if ctx != ""  →  FALSE
          ↓
systemPrompt = basePrompt တစ်ခုတည်း
          ↓
LLM ထံ ပေးပို့  (RAG context မပါဘဲ)
```

---

### 🧠 LLM ဘာပြုလုပ်မလဲ

Qdrant data မရှိရင် LLM သည် **သူ့ကိုယ်ပိုင် pre-trained knowledge** ကိုသာ သုံးပြီး ဖြေတယ်

```
basePrompt:  "You are Frontiir AI Assistant..."
                       ↓
  LLM ထဲမှာ သိပြီးသား Myanmar ISP knowledge
  + general internet troubleshooting knowledge
                       ↓
ဖြေဆိုတယ် — accuracy နည်းနိုင်သော်လည်း ဖြေနိုင်သေးတယ်
```

---

### 📊 Qdrant data ရှိ/မရှိ နှိုင်းယှဉ်ချက်

| အခြေအနေ | LLM ဖြေပုံ | Accuracy |
|---|---|---|
| Qdrant data **ရှိ** ✅ | `"Frontiir Home Package သည် 5,000 ကျပ်မှ..."` | ⭐⭐⭐⭐⭐ တိကျ |
| Qdrant data **မရှိ** ❌ | `"Frontiir ၏ package များသည် ယေဘုယျအားဖြင့်..."` | ⭐⭐ မတိကျနိုင် |

---

### 💡 အရေးကြီးတဲ့ နားလည်မှု

```
Qdrant data = Extra study notes
LLM         = Student ကိုယ်တိုင်

notes ရှိ  → notes ကြည့်ပြီး ဖြေ  (ပိုတိကျ)
notes မရှိ → ကိုယ်သိတာနဲ့ ဖြေ   (ကိုယ်တိုင် knowledge)
```

> ✅ **Production အတွက်:** JSONL training data တွေကို ingest လုပ်ထားဖို့ အရေးကြီးတယ်
> LLM ရဲ့ accuracy သည် **Qdrant ထဲ data ပေါ်မှာ တိုက်ရိုက် မူတည်တယ်**

---

## 🔧 Message Build လုပ်ပုံ

```
┌─────────────────────────────────────────────┐
│  messages array (LLM ကို ပို့တဲ့ order)      │
├─────────────────────────────────────────────┤
│  [0] system  → basePrompt + context         │
│  [1] user    → (memory) ယခင် user msg       │
│  [2] assist  → (memory) ယခင် AI response    │
│  [3] user    → (memory) user msg            │
│  [4] assist  → (memory) AI response         │
│  ...                                        │
│  [N] user    → ယခု မေးခွန်း (current)       │
└─────────────────────────────────────────────┘
```

> Memory သည် နောက်ဆုံး **10 messages** သာ သိမ်းတယ် (`memory.go`)

---

*Last updated: 2026-03-04 | Frontiir AI Project*
