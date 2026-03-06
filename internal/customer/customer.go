// internal/customer/customer.go
// Detects Myanmar phone numbers from user messages and fetches customer data from Frontiir API.
package customer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// phonePattern matches Myanmar phone numbers:
// 09-12345678, 09 12345678, 0912345678, +959-12345678, +95912345678
var phonePattern = regexp.MustCompile(`(?:\+?95[-\s]?|0)(9\d{7,9})`)

// APIResponse is the expected shape from the Frontiir Customer API
type APIResponse struct {
	CustomerID     string `json:"customer_id"`
	Name           string `json:"name"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	Location       string `json:"location"`
	Package        string `json:"package"`
	Status         string `json:"status"`
	Expiry         string `json:"expiry"`
	Balance        string `json:"balance"`
	PaymentHistory string `json:"payment_history"`
	Message        string `json:"message"`
}

// ExtractPhone returns the first Myanmar phone number found in text (normalized to 09XXXXXXX), or empty string.
func ExtractPhone(text string) string {
	m := phonePattern.FindStringSubmatch(text)
	if len(m) >= 2 {
		return "0" + m[1] // e.g. "09-12345678" → "09" + "12345678"
	}
	return ""
}

// FetchInfo calls the Customer API by phone number and returns a formatted context string.
func FetchInfo(phone string) string {
	apiURL := os.Getenv("CUSTOMER_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:9000/api/customer"
	}

	url := fmt.Sprintf("%s/%s", strings.TrimRight(apiURL, "/"), phone)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Sprintf("[Customer API Error] Could not reach Customer API for %s: %v", phone, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("[Customer API Error] Failed to read response for %s", phone)
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Sprintf("[Customer Info] ဖုန်းနံပါတ် %s နှင့် ဆက်သွယ်ထားသော customer မတွေ့ပါ။ ဖုန်းနံပါတ် မှန်ကန်မှု စစ်ဆေးပါ။", phone)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("[Customer API Error] Status %d for phone %s: %s", resp.StatusCode, phone, string(body))
	}

	var data APIResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Sprintf("[Customer Info for %s]\n%s", phone, string(body))
	}

	return formatResponse(phone, data)
}

func formatResponse(phone string, d APIResponse) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[Customer Info for %s]\n", phone))
	if d.CustomerID != "" {
		sb.WriteString(fmt.Sprintf("  Customer ID    : %s\n", d.CustomerID))
	}
	if d.Name != "" {
		sb.WriteString(fmt.Sprintf("  Name           : %s\n", d.Name))
	}
	if d.Phone != "" {
		sb.WriteString(fmt.Sprintf("  Phone          : %s\n", d.Phone))
	}
	if d.Package != "" {
		sb.WriteString(fmt.Sprintf("  Package        : %s\n", d.Package))
	}
	if d.Status != "" {
		sb.WriteString(fmt.Sprintf("  Status         : %s\n", d.Status))
	}
	if d.Expiry != "" {
		sb.WriteString(fmt.Sprintf("  Expiry         : %s\n", d.Expiry))
	}
	if d.Balance != "" {
		sb.WriteString(fmt.Sprintf("  Balance        : %s\n", d.Balance))
	}
	if d.PaymentHistory != "" {
		sb.WriteString(fmt.Sprintf("  Payment History: %s\n", d.PaymentHistory))
	}
	if d.Address != "" {
		sb.WriteString(fmt.Sprintf("  Address        : %s\n", d.Address))
	}
	if d.Location != "" {
		sb.WriteString(fmt.Sprintf("  Location       : %s\n", d.Location))
	}
	if d.Message != "" {
		sb.WriteString(fmt.Sprintf("  Note           : %s\n", d.Message))
	}
	return sb.String()
}
