// internal/cpe/cpe.go
// Detects CPE ID from user message and fetches CPE data from Frontiir API.
package cpe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// cpePattern matches CPE IDs like: CPE-123456, CPE-ABC123, cpe-001234
var cpePattern = regexp.MustCompile(`(?i)\bCPE[-_]?([A-Z0-9]{4,12})\b`)

// APIResponse is the expected shape from the Frontiir CPE API
type APIResponse struct {
	CPEID      string `json:"cpe_id"`
	Status     string `json:"status"`
	Signal     string `json:"signal"`
	Uptime     string `json:"uptime"`
	IPAddress  string `json:"ip_address"`
	Location   string `json:"location"`
	Message    string `json:"message"`
}

// ExtractID returns the first CPE ID found in text, or empty string.
func ExtractID(text string) string {
	m := cpePattern.FindStringSubmatch(text)
	if len(m) >= 1 {
		return strings.ToUpper(m[0])
	}
	return ""
}

// FetchInfo calls the CPE API and returns a formatted context string.
func FetchInfo(cpeID string) string {
	apiURL := os.Getenv("CPE_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:9000/api/cpe"
	}

	url := fmt.Sprintf("%s/%s", strings.TrimRight(apiURL, "/"), cpeID)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("[CPE API Error] Could not reach CPE API for %s: %v", cpeID, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("[CPE API Error] Failed to read response for %s", cpeID)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Sprintf("[CPE Info] CPE ID %s မတွေ့ပါ။ CPE ID မှန်ကန်မှု စစ်ဆေးပါ။", cpeID)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("[CPE API Error] Status %d for CPE %s: %s", resp.StatusCode, cpeID, string(body))
	}

	var data APIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		// API returned non-JSON — return raw text
		return fmt.Sprintf("[CPE Info for %s]\n%s", cpeID, string(body))
	}

	return formatResponse(cpeID, data)
}

func formatResponse(cpeID string, d APIResponse) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[CPE Info for %s]\n", cpeID))
	if d.Status != "" {
		sb.WriteString(fmt.Sprintf("  Status    : %s\n", d.Status))
	}
	if d.Signal != "" {
		sb.WriteString(fmt.Sprintf("  Signal    : %s\n", d.Signal))
	}
	if d.Uptime != "" {
		sb.WriteString(fmt.Sprintf("  Uptime    : %s\n", d.Uptime))
	}
	if d.IPAddress != "" {
		sb.WriteString(fmt.Sprintf("  IP Address: %s\n", d.IPAddress))
	}
	if d.Location != "" {
		sb.WriteString(fmt.Sprintf("  Location  : %s\n", d.Location))
	}
	if d.Message != "" {
		sb.WriteString(fmt.Sprintf("  Note      : %s\n", d.Message))
	}
	return sb.String()
}
