# 🔄 Frontiir AI Chat — Sequence Diagram

> **မြန်မာဘာသာ** — Project Flow Reference
> Full chat request lifecycle from Browser to LLM and back.

---

## 📊 Sequence Diagram

```mermaid
sequenceDiagram
    actor User as 👤 User (Browser)
    participant Server as 🌐 Server<br/>(server.go)
    participant Guard as 🛡️ Guard<br/>(guard.go)
    participant Agent as 🤖 Agent<br/>(agent.go)
    participant Embedder as 🔢 Embedder<br/>(embedder.go)
    participant Qdrant as 🗄️ Qdrant<br/>(vectorstore.go)
    participant CPE as 📡 CPE API<br/>(cpe.go)
    participant Customer as 👥 Customer API<br/>(customer.go)
    participant Memory as 💾 Memory<br/>(memory.go)
    participant LLM as 🧠 Ollama LLM<br/>(ollama.go)
    participant Cleaner as 🧹 Cleaner<br/>(postprocess.go)

    User->>Server: POST /api/chat { message }

    Note over Server: Request decode လုပ်တယ်

    Server->>Guard: IsDenied(message)
    alt rude keyword ပါနေရင်
        Guard-->>Server: true
        Server-->>User: ❌ ယဉ်ကျေးစွာ ငြင်းဆိုသော response
    else ပုံမှန် message
        Guard-->>Server: false

        Server->>Agent: ChatStream(message, onToken)

        Agent->>Embedder: Embed(message)
        Embedder->>LLM: POST /api/embeddings (nomic-embed-text)
        LLM-->>Embedder: vector [0.23, 0.81, ...]
        Embedder-->>Agent: queryVector

        Agent->>Qdrant: Search(collection, queryVector, topK=3)
        Qdrant-->>Agent: docs[] (ဆင်တူသော documents 3 ခု)

        Agent->>Memory: Get()
        Memory-->>Agent: conversation history[]

        alt CPE-XXXXXX pattern ပါနေရင်
            Agent->>CPE: ExtractID(message) → cpeID
            Agent->>CPE: FetchInfo(cpeID)
            CPE->>CPE: GET {CPE_API_URL}/{cpeID}
            CPE-->>Agent: [CPE Info] status, signal, uptime...
        end

        alt 09-xxxxxxx phone pattern ပါနေရင်
            Agent->>Customer: ExtractPhone(message) → phone
            Agent->>Customer: FetchInfo(phone)
            Customer->>Customer: GET {CUSTOMER_API_URL}/{phone}
            Customer-->>Agent: [Customer Info] name, package, balance...
        end

        Note over Agent: buildContext()<br/>docs + CPE data + Customer data ပေါင်း

        Agent->>LLM: ChatStream(messages, onToken)
        Note over LLM: system prompt + memory +<br/>context + user message

        loop token တစ်ခုချင်း stream
            LLM-->>Agent: token
            Agent->>Cleaner: Clean(token)
            Cleaner-->>Agent: cleaned token
            Agent-->>Server: onToken(cleaned)
            Server-->>User: SSE data: {"token": "..."}
        end

        Agent->>Memory: Add("user", message)
        Agent->>Memory: Add("assistant", fullResponse)

        Server-->>User: SSE data: {"done": "true"}
    end
```

---

## 📝 တစ်ဆင့်ချင်း ရှင်းလင်းချက် (မြန်မာဘာသာ)

### 1️⃣ User → Server

Browser မှ `POST /api/chat` ကို JSON body နဲ့ ပေးပို့တယ်

---

### 2️⃣ Guard စစ်ဆေးမှု

`keywords.json` ထဲက blocked words တွေနဲ့ compare လုပ်တယ်

- ပါရင် → LLM ကို မပေးဘဲ ချက်ချင်း ငြင်းတယ်
- မပါရင် → ဆက်သွားတယ်

---

### 3️⃣ Embed + Search

Message ကို vector ပြောင်းပြီး Qdrant မှာ ဆင်တူတဲ့ knowledge base docs ၃ ခု ရှာတယ်

---

### 4️⃣ Tool Checks

| Detect | Tool | Data |
|---|---|---|
| `CPE-XXXXXX` | CPE API | signal, status, uptime |
| `09-xxxxxxx` | Customer API | name, package, balance |

---

### 5️⃣ Context Build

RAG docs + CPE data + Customer data ကို ပေါင်းပြီး LLM ကို context အဖြစ် ပေးတယ်

---

### 6️⃣ Stream

LLM က token တစ်ခုချင်း ထုတ်ပေးတဲ့အချိန် → Cleaner မှာ မြန်မာ text fix → Browser ကို SSE နဲ့ real-time ပို့တယ်

---

### 7️⃣ Memory Save

Conversation ကို memory ထဲ သိမ်းတယ် (နောက် message မှာ context သိအောင်)

---

*Last updated: 2026-03-04 | Frontiir AI Project*
