// cmd/setup/main.go
// Automates the same steps as colab_ollama.ipynb in Go.
// Run this once inside a Linux machine (e.g. Google Colab) to prepare the environment.
//
// Usage:
//
//	go run ./cmd/setup --ngrok-token=YOUR_TOKEN --model=yxchia/seallms-v3-7b:Q4_K_M
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	ngrokToken = flag.String("ngrok-token", "", "ngrok authtoken (required)")
	model      = flag.String("model", "yxchia/seallms-v3-7b:Q4_K_M", "Ollama chat model to pull")
	embedModel = flag.String("embed-model", "nomic-embed-text", "Ollama embedding model to pull")
	ollamaHost = flag.String("ollama-host", "0.0.0.0:11434", "Ollama listen address")
)

func main() {
	flag.Parse()

	step(1, "Check GPU")
	checkGPU()

	step(2, "Install zstd")
	must(run("apt-get", "install", "-y", "zstd"))

	step(3, "Install Ollama")
	installOllama()

	step(4, "Start Ollama server")
	startOllama()

	step(5, "Pull chat model: "+*model)
	must(run("ollama", "pull", *model))

	step(6, "Pull embed model: "+*embedModel)
	must(run("ollama", "pull", *embedModel))

	step(7, "Test Ollama chat")
	testOllama()

	if *ngrokToken != "" {
		step(8, "Start ngrok tunnel")
		startNgrok()
	} else {
		fmt.Println("⚠️  --ngrok-token not set, skipping ngrok.")
		fmt.Printf("   Ollama is available at http://localhost:11434\n")
	}

	step(9, "Keep alive")
	keepAlive()
}

// ── Step helpers ──────────────────────────────────────────────────────────────

func step(n int, name string) {
	fmt.Printf("\n\033[1;36m=== Step %d — %s ===\033[0m\n", n, name)
}

// ── Step 1: Check GPU ─────────────────────────────────────────────────────────

func checkGPU() {
	out, err := exec.Command("nvidia-smi").CombinedOutput()
	if err != nil {
		fmt.Println("⚠️  No NVIDIA GPU detected — Ollama will run on CPU.")
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "Tesla") || strings.Contains(line, "NVIDIA") {
			fmt.Println("✅ GPU found:", strings.TrimSpace(line))
			return
		}
	}
	fmt.Println("✅ GPU info:\n", string(out))
}

// ── Step 3: Install Ollama ────────────────────────────────────────────────────

func installOllama() {
	// Download and pipe the install script to sh
	cmd := exec.Command("sh", "-c", "curl -fsSL https://ollama.com/install.sh | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("❌ Ollama install failed:", err)
		os.Exit(1)
	}
	fmt.Println("✅ Ollama installed.")
}

// ── Step 4: Start Ollama server ───────────────────────────────────────────────

func startOllama() {
	cmd := exec.Command("ollama", "serve")
	cmd.Env = append(os.Environ(), "OLLAMA_HOST="+*ollamaHost)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		fmt.Println("❌ Could not start Ollama:", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Ollama server started (PID %d), waiting for it to be ready...\n", cmd.Process.Pid)

	// Wait until /api/tags responds
	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		resp, err := http.Get("http://localhost:11434/api/tags")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("✅ Ollama is ready.")
			return
		}
	}
	fmt.Println("⚠️  Ollama did not respond after 20s — continuing anyway.")
}

// ── Step 7: Test ──────────────────────────────────────────────────────────────

func testOllama() {
	body, _ := json.Marshal(map[string]any{
		"model":    *model,
		"messages": []map[string]string{{"role": "user", "content": "Say hi in one sentence."}},
		"stream":   false,
	})
	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("❌ Test failed:", err)
		return
	}
	defer resp.Body.Close()

	var result struct {
		Message struct{ Content string } `json:"message"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println("✅ Model replied:", result.Message.Content)
}

// ── Step 8: ngrok tunnel ──────────────────────────────────────────────────────

func startNgrok() {
	// Install pyngrok via pip then call Python to open the tunnel
	exec.Command("pip", "install", "pyngrok", "-q").Run()

	script := fmt.Sprintf(`
from pyngrok import ngrok
import time
ngrok.set_auth_token("%s")
tunnel = ngrok.connect(11434)
url = tunnel.public_url
print(url)
`, *ngrokToken)

	cmd := exec.Command("python3", "-c", script)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("❌ ngrok failed:", err)
		return
	}

	ollamaURL := strings.TrimSpace(string(out))
	fmt.Println("✅ Ollama is public!")
	fmt.Println()
	fmt.Println("  OLLAMA_URL =", ollamaURL)
	fmt.Println()
	fmt.Println("  Run your Go app with:")
	fmt.Printf("  OLLAMA_URL=%s go run .\n", ollamaURL)
}

// ── Step 9: Keep alive ────────────────────────────────────────────────────────

func keepAlive() {
	fmt.Println("🔄 Keeping process alive... (Ctrl+C to stop)")
	scanner := bufio.NewScanner(os.Stdin)
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			fmt.Print(".")
		default:
			if scanner.Scan() {
				// allow typing 'quit' to stop
				if strings.TrimSpace(scanner.Text()) == "quit" {
					return
				}
			}
		}
	}
}

// ── Utilities ─────────────────────────────────────────────────────────────────

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func must(err error) {
	if err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}

// Colab Terminal မှာ:

// # ngrok token ဖြင့် run မည်
// go run ./cmd/setup --ngrok-token=3ACOWs1OOGjQbbWPwDdtd8od9HC_4Hz3j5VwiEvRPPfEETLUC

// # model ပြောင်းချင်ရင်
// go run ./cmd/setup \
//   --ngrok-token=3ACOWs1OOGjQbbWPwDdtd8od9HC_4Hz3j5VwiEvRPPfEETLUC \
//   --model=qwen2.5:3b

// Output ဒီလို ထွက်မည်:

// === Step 1 — Check GPU ===
// ✅ GPU found: Tesla T4

// === Step 4 — Start Ollama server ===
// ✅ Ollama is ready.

// === Step 8 — Start ngrok tunnel ===
// ✅ Ollama is public!

//   OLLAMA_URL = https://xxxx.ngrok-free.app

//   Run your Go app with:
//   OLLAMA_URL=https://xxxx.ngrok-free.app go run .
