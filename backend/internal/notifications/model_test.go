package notifications

import (
	"encoding/json"
	"testing"
)

func TestUnreadNotificationSerializesReadAtAsNull(t *testing.T) {
	payload, err := json.Marshal(Notification{ID: "notification-1"})
	if err != nil {
		t.Fatalf("marshal notification: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal notification: %v", err)
	}
	readAt, exists := decoded["read_at"]
	if !exists {
		t.Fatal("read_at is missing from unread notification payload")
	}
	if readAt != nil {
		t.Fatalf("read_at = %v, want null", readAt)
	}
}
