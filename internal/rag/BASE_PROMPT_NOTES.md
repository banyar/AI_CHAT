# 🟡 basePrompt ဆိုတာ ဘာလဲ၊ ဘယ်လိုအလုပ်လုပ်လဲ

> File: `internal/rag/agent.go` — `const basePrompt`

---

## ❓ ဘာလဲ? (What)

**basePrompt** ဆိုတာ LLM ကို စတင်မှာကြားတဲ့ **"သင်ကြားပေးချက်"** ဖြစ်တယ်။

Human မှာ employee ကို `"သင်တစ်ဦးသည် Frontiir customer support ဖြစ်သည်..."` လို့ briefing ပေးသလိုပဲ၊
LLM ကိုလည်း chat စတင်တိုင်း ဒီ instruction ကို ပေးတယ်။

---

## ⏱️ ဘယ်အချိန် အလုပ်လုပ်လဲ? (When)

Chat တိုင်း — user message **တစ်ကြိမ်** ပေးပို့တဲ့အချိန်မှာ:

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

`agent.go:69-75` မှာ ဒီလို build လုပ်တယ်:

```go
systemPrompt := basePrompt          // အမြဲ basePrompt နဲ့ စတယ်

if ctx := buildContext(...); ctx != "" {
    systemPrompt = basePrompt + "\n\nContext:\n" + ctx  // RAG/CPE/Customer data ထပ်တယ်
}

messages := []Message{
    {Role: "system",  Content: systemPrompt},  // ← basePrompt ဒီမှာ
    // ...memory history...
    {Role: "user",    Content: userMsg},
}
```

---

## 🔍 basePrompt ရဲ့ အပိုင်းတစ်ခုချင်း ရှင်းလင်းချက်

---

### 🔵 အပိုင်း ၁ — Role သတ်မှတ်ချက်

```
You are Frontiir AI Assistant — a helpful customer support
assistant for Frontiir, Myanmar's leading fiber internet provider.
```

**ဘာကြောင့်:** LLM ကို `"ဘယ်သူလဲ"` သတ်မှတ်ပေးတယ်။

| | မသတ်မှတ်ရင် | သတ်မှတ်ရင် |
|---|---|---|
| AI role | General assistant | Frontiir Support |
| Topic | ဘာမဆို ဖြေတယ် | Frontiir topic ပဲ |

---

### 🟢 အပိုင်း ၂ — ဘာသာစကား Rule

```
- If the user writes in Burmese (မြန်မာ), reply ONLY in pure Burmese.
- If the user writes in English, reply ONLY in English.
```

**ဘာကြောင့်:** LLM တွေသည် automatically ဘာသာပေါင်းစပ်တတ်တယ်။

**မရှိရင်** ❌
```
User:  "internet မကောင်းဘူး"
AI:    "Your internet might be experiencing issues.
        ကျေးဇူးပြု၍ router ကို restart please do it."  ← ရောနေတယ်
```

**ရှိရင်** ✅
```
User:  "internet မကောင်းဘူး"
AI:    "Router ကို restart လုပ်ပြီး ၅ မိနစ် စောင့်ပါ။"  ← သန့်တယ်
```

---

### 🟠 အပိုင်း ၃ — Mixed Word တားမြစ်ချက်

```
- Never combine two languages in one word.
  Write "ဥပမာ" OR "example" — never "ဥပမable".
```

**ဘာကြောင့်:** LLM တွေသည် တစ်ခါတရံ word ကိုယ်တိုင် ရောနေတဲ့ artifact ထုတ်တယ်။

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

**ဘာကြောင့်:** `"CPE"`, `"router"`, `"ONT"` ဆိုတာတွေကို မြန်မာလို ဘာသာပြန်ရင် ဆိုးလာမယ်။

| Rule မရှိရင် ❌ | Rule ရှိရင် ✅ |
|---|---|
| `"သင့် ဖောက်သည် နေရာ ပစ္စည်း မကောင်းဘူး"` | `"သင့် CPE device မကောင်းဘူး"` |

---

### 🟣 အပိုင်း ၅ — Example Format

```
- Use "ဥပမာ -" (not "ဥပမာ:" or "example:") when giving examples
```

**ဘာကြောင့်:** Output format ကို **consistent** ထားဖို့ — response တွေ ညီတူ ပုံစံ ဖြစ်အောင်

---

## 📊 Summary — basePrompt ရဲ့ Role

```
basePrompt = AI ရဲ့ "Job Description" + "လုပ်ပိုင်ခွင့် စည်းကမ်း"
                          ↓
         Chat တိုင်း System Message အနေနဲ့ LLM ကို ပေးတယ်
                          ↓
    LLM က "Frontiir Customer Support" role ဆောင်ပြီး ဖြေတယ်
```

| အချက် | မရှိရင် ❌ | ရှိရင် ✅ |
|---|---|---|
| Topic | ဘာ topic မဆို ဖြေတယ် | Frontiir topic ပဲ |
| ဘာသာ | ရောနေတယ် | သန့်တယ် |
| Mixed words | ဖြစ်နိုင်တယ် | မဖြစ်တော့ |
| Format | မညီ | ညီ |

---

*Last updated: 2026-03-04 | Frontiir AI Project*
