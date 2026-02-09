package ssh

import (
	"context"
	"testing"
	"time"
)

func TestConnectivityInvalidAddress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := TestConnectivity(ctx, ConnectivityTestRequest{
		Address:  "192.0.2.1", // RFC 5737 TEST-NET, guaranteed unreachable
		Port:     22,
		Username: "test",
		Password: "test",
	})

	if result.Reachable {
		t.Error("expected Reachable=false for unreachable address")
	}
	if result.Error == "" {
		t.Error("expected non-empty error")
	}
}

func TestConnectivityMissingKeyFile(t *testing.T) {
	result := TestConnectivity(context.Background(), ConnectivityTestRequest{
		Address:        "localhost",
		Port:           22,
		Username:       "test",
		PrivateKeyPath: "/nonexistent/path/id_rsa",
	})

	if result.Reachable {
		t.Error("expected Reachable=false when key file missing")
	}
	if result.Error == "" {
		t.Error("expected non-empty error for missing key file")
	}
}

func TestConnectivityDefaultPort(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := TestConnectivity(ctx, ConnectivityTestRequest{
		Address:  "192.0.2.1",
		Username: "test",
		Password: "test",
		// Port left at 0, should default to 22
	})

	// Just verifying it doesn't panic with port 0
	if result.Reachable {
		t.Error("expected Reachable=false for unreachable address")
	}
}
