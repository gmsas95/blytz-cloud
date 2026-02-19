package caddy

import (
	"testing"
)

func TestClient(t *testing.T) {
	client := NewClient("http://localhost:2019")

	if client.adminURL != "http://localhost:2019" {
		t.Errorf("Expected admin URL http://localhost:2019, got %s", client.adminURL)
	}
}
