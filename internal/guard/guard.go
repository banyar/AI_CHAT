// internal/guard/guard.go
// Blocks rude or inappropriate messages before they reach the LLM.
// Keywords are loaded from keywords.json — edit that file to add/remove words.
package guard

import (
	"encoding/json"
	_ "embed"
	"log"
	"strings"
)

//go:embed keywords.json
var keywordsFile []byte

var blockedKeywords []string
var warnedKeywords []string

func init() {
	var cfg struct {
		BlockedKeywords []string `json:"blocked_keywords"`
		WarnedKeywords  []string `json:"warned_keywords"`
	}
	if err := json.Unmarshal(keywordsFile, &cfg); err != nil {
		log.Printf("[guard] Warning: failed to load keywords.json: %v", err)
		return
	}
	blockedKeywords = cfg.BlockedKeywords
	warnedKeywords = cfg.WarnedKeywords
	log.Printf("[guard] Loaded %d blocked, %d warned keywords from keywords.json",
		len(blockedKeywords), len(warnedKeywords))
}

// deniedBurmese is the polite rejection response in Burmese.
const deniedBurmese = "ကျွန်တော်သည် Frontiir AI Assistant ဖြစ်ပါသည်။ ယဉ်ကျေးသော ဘာသာစကားဖြင့် ပြောဆိုပါရန် မေတ္တာရပ်ခံပါသည်။ Frontiir ဝန်ဆောင်မှုများနှင့် ပတ်သက်၍ မည်သည့်အကူအညီမဆို ပေးနိုင်ပါသည်။"

// deniedEnglish is the polite rejection response in English.
const deniedEnglish = "I'm Frontiir AI Assistant. Please use respectful language so I can assist you better. I'm here to help with any Frontiir service questions."

// IsDenied returns true if the message contains any blocked keyword.
func IsDenied(message string) bool {
	lower := strings.ToLower(message)
	for _, kw := range blockedKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// DeniedResponse returns the appropriate polite rejection message
// based on whether the user wrote in Burmese or English.
func DeniedResponse(message string) string {
	if looksLikeBurmese(message) {
		return deniedBurmese
	}
	return deniedEnglish
}

// IsWarned returns true if the message contains a warned (soft) keyword.
func IsWarned(message string) bool {
	lower := strings.ToLower(message)
	for _, kw := range warnedKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// WarnPrefix returns a polite warning to prepend before the LLM answer.
func WarnPrefix(message string) string {
	if looksLikeBurmese(message) {
		return "⚠️ ယဉ်ကျေးသော ဘာသာစကားဖြင့် ပြောဆိုပါရန် မေတ္တာရပ်ခံပါသည်။\n"
	}
	return "⚠️ Please use respectful language.\n"
}

// looksLikeBurmese checks if the message contains Myanmar Unicode characters.
func looksLikeBurmese(text string) bool {
	for _, r := range text {
		if r >= 0x1000 && r <= 0x109F {
			return true
		}
	}
	return false
}
