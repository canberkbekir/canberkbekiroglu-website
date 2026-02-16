package analytics

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type ipAPIResponse struct {
	Status string `json:"status"`
	City   string `json:"city"`
}

// FetchCityFromAPI resolves an IP address to a city using ip-api.com.
// Returns "Local" for loopback addresses, "Unknown" on any error.
func FetchCityFromAPI(ip string) string {
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return "Local"
	}

	// Check if it's a private/reserved IP.
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() {
		return "Local"
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://ip-api.com/json/%s?fields=status,city", ip))
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()

	var result ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "Unknown"
	}

	if result.Status != "success" || result.City == "" {
		return "Unknown"
	}
	return result.City
}
