package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// parsePublishedAt parses a user-supplied year ("2001") or full date ("2001-09-01")
// into a time.Time. Returns an error if the string is not a recognised format.
func parsePublishedAt(s string) (time.Time, error) {
	// Try full date first
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	// Try year only
	if t, err := time.Parse("2006", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("unrecognised date format: %q", s)
}
