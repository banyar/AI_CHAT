#!/bin/bash
# ============================================================
# Frontiir AI — Load fine-tuned GGUF model into Ollama & Run App
# Usage:
#   bash train.sh                          → base model + start app
#   bash train.sh /path/to/file.gguf      → fine-tuned model + start app
#   bash train.sh --setup-only            → base model, NO app start
#   bash train.sh /path/to/file.gguf --setup-only
# ============================================================

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
MODEL_NAME="frontiir-ai"
MODELFILE="$PROJECT_DIR/Modelfile"

# ── Parse arguments ──
GGUF_FILE=""
SETUP_ONLY=false
for arg in "$@"; do
    case "$arg" in
        --setup-only) SETUP_ONLY=true ;;
        *)            GGUF_FILE="$arg" ;;
    esac
done

echo "=============================="
echo "  Frontiir AI — Ollama Setup"
echo "=============================="

# Check ollama is installed
if ! command -v ollama &>/dev/null; then
    echo "❌ Ollama is not installed."
    echo "   Install: curl -fsSL https://ollama.com/install.sh | sh"
    exit 1
fi

# ── Mode 1: GGUF file provided (after Colab fine-tuning) ──
if [ -n "$GGUF_FILE" ]; then
    if [ ! -f "$GGUF_FILE" ]; then
        echo "❌ GGUF file not found: $GGUF_FILE"
        exit 1
    fi

    echo ""
    echo "📦 GGUF file: $GGUF_FILE"
    echo "🔨 Creating Modelfile with GGUF..."

    # Build a temporary Modelfile pointing to the GGUF
    TMP_MODELFILE=$(mktemp)
    cat > "$TMP_MODELFILE" <<EOF
FROM $GGUF_FILE

SYSTEM """
You are Frontiir AI Assistant — a helpful customer support assistant for Frontiir, Myanmar's leading fiber internet provider.

You answer questions about:
- Internet packages (Home, Business, Enterprise)
- Speeds, coverage, installation, and contracts
- Billing, payments (KBZPay, Wave Money, AYA Pay, CBPay)
- Technical troubleshooting (router, WiFi, speed issues)
- Frontiir app usage
- Business solutions (Dedicated Internet, MPLS, VPN, CCTV)
- Company information and careers

Rules:
- Always be polite and helpful
- Answer in the same language the user uses (Burmese or English)
- For complex issues, direct users to call 01-234567 or email support@frontiir.com
- Do not make up information — if unsure, ask the user to contact customer service
"""

PARAMETER temperature 0.3
PARAMETER top_p 0.9
PARAMETER repeat_penalty 1.1
EOF

    echo "🚀 Creating Ollama model '$MODEL_NAME' from fine-tuned GGUF..."
    ollama create "$MODEL_NAME" -f "$TMP_MODELFILE"
    rm "$TMP_MODELFILE"

# ── Mode 2: No GGUF — use base qwen2.5:3b with system prompt ──
else
    echo ""
    echo "ℹ️  No GGUF file provided."
    echo "   Using base qwen2.5:3b + Frontiir system prompt (Modelfile)."
    echo ""

    # Pull base model if not present
    if ! ollama list | grep -q "qwen2.5:3b"; then
        echo "⬇️  Pulling qwen2.5:3b base model (~2GB)..."
        ollama pull qwen2.5:3b
    else
        echo "✅ qwen2.5:3b already available"
    fi

    echo "🚀 Creating Ollama model '$MODEL_NAME' from Modelfile..."
    ollama create "$MODEL_NAME" -f "$MODELFILE"
fi

# ── Verify ──
echo ""
if ! ollama list | grep -q "$MODEL_NAME"; then
    echo "❌ Model creation failed. Check the error above."
    exit 1
fi

echo "✅ Model '$MODEL_NAME' created successfully!"
echo ""

# ── Setup-only mode: just print instructions ──
if [ "$SETUP_ONLY" = true ]; then
    echo "=============================="
    echo "  Setup complete. To run:"
    echo "=============================="
    echo "  ollama run $MODEL_NAME"
    echo "  OLLAMA_MODEL=$MODEL_NAME go run ."
    exit 0
fi

# ── Auto-start the Go app ──
echo "=============================="
echo "  Starting Frontiir AI App"
echo "=============================="
echo ""

# Check Go is installed
if ! command -v go &>/dev/null; then
    echo "❌ Go is not installed. Install from https://go.dev/dl/"
    echo "   Then run manually: OLLAMA_MODEL=$MODEL_NAME go run ."
    exit 1
fi

# Check ollama is serving (start if not)
if ! curl -sf http://localhost:11434 &>/dev/null; then
    echo "⚙️  Starting Ollama server..."
    ollama serve &
    sleep 3
fi

echo "🚀 Running: OLLAMA_MODEL=$MODEL_NAME go run ."
echo ""
cd "$PROJECT_DIR" && OLLAMA_MODEL="$MODEL_NAME" go run .
