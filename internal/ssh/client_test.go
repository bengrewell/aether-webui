package ssh

import (
	"net"
	"testing"
)

func TestHostPortDefault(t *testing.T) {
	// Verify that a bare host gets :22 appended.
	host := "10.0.0.1"
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "22")
	}
	if host != "10.0.0.1:22" {
		t.Fatalf("expected 10.0.0.1:22, got %s", host)
	}

	// Verify that host:port is preserved.
	host = "10.0.0.1:2222"
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "22")
	}
	if host != "10.0.0.1:2222" {
		t.Fatalf("expected 10.0.0.1:2222, got %s", host)
	}
}

func TestDialNoAuth(t *testing.T) {
	_, err := Dial(t.Context(), Config{
		Host: "127.0.0.1:22",
		User: "test",
	})
	if err == nil {
		t.Fatal("expected error with no auth method")
	}
}
