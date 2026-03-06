// internal/postprocess/cleaner.go
// Cleans model output artifacts — fixes Burmese-English token mixing.
package postprocess

import (
	"regexp"
	"strings"
)

// mixed-word patterns the model tends to generate
var mixedPatterns = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	// "ဥပမable"  → "ဥပမာ"
	{regexp.MustCompile(`ဥပမ\w*able\w*`), "ဥပမာ"},
	// "ဥပမာ:" with no space → "ဥပမာ -"
	{regexp.MustCompile(`ဥပမာ\s*:`), "ဥပမာ -"},
	// "ဥပမa" variants
	{regexp.MustCompile(`ဥပမ[a-zA-Z]+`), "ဥပမာ"},
	// "example:" at start of Burmese sentence → "ဥပမာ -"
	{regexp.MustCompile(`(?i)\bexample\s*:`), "ဥပမာ -"},
	// Broken Burmese endings mixed with English
	{regexp.MustCompile(`ပါသ[a-zA-Z]+`), "ပါသည်"},
	{regexp.MustCompile(`ဖြစ်ပါ[a-zA-Z]+`), "ဖြစ်ပါသည်"},
	{regexp.MustCompile(`တွင်[a-zA-Z]+`), "တွင်"},
	// Double spaces
	{regexp.MustCompile(`  +`), " "},
	// "，" (Chinese comma) → Burmese-friendly "၊"
	{regexp.MustCompile(`，`), "၊ "},
}

// Clean applies all fix patterns to the model output.
func Clean(text string) string {
	for _, p := range mixedPatterns {
		text = p.pattern.ReplaceAllString(text, p.replacement)
	}
	return strings.TrimSpace(text)
}
