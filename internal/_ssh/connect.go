package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

// ConnectivityTestRequest describes the parameters for an SSH connectivity test.
type ConnectivityTestRequest struct {
	Address        string
	Port           int
	Username       string
	Password       string
	PrivateKeyPath string
}

// ConnectivityTestResult holds the outcome of an SSH connectivity test.
type ConnectivityTestResult struct {
	Reachable     bool   `json:"reachable"`
	Authenticated bool   `json:"authenticated"`
	LatencyMs     int64  `json:"latency_ms"`
	Error         string `json:"error,omitempty"`
	ServerVersion string `json:"server_version,omitempty"`
}

const defaultTimeout = 10 * time.Second

// TestConnectivity tests SSH connectivity to a remote host.
// It performs a TCP dial followed by an SSH handshake.
func TestConnectivity(ctx context.Context, req ConnectivityTestRequest) *ConnectivityTestResult {
	result := &ConnectivityTestResult{}

	if req.Port == 0 {
		req.Port = 22
	}

	addr := fmt.Sprintf("%s:%d", req.Address, req.Port)

	// Build auth methods
	var authMethods []gossh.AuthMethod
	if req.PrivateKeyPath != "" {
		keyData, err := os.ReadFile(req.PrivateKeyPath)
		if err != nil {
			result.Error = fmt.Sprintf("failed to read private key: %v", err)
			return result
		}
		signer, err := gossh.ParsePrivateKey(keyData)
		if err != nil {
			result.Error = fmt.Sprintf("failed to parse private key: %v", err)
			return result
		}
		authMethods = append(authMethods, gossh.PublicKeys(signer))
	}
	if req.Password != "" {
		authMethods = append(authMethods, gossh.Password(req.Password))
	}

	config := &gossh.ClientConfig{
		User: req.Username,
		Auth: authMethods,
		// Note: production deployments should use known_hosts verification
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         defaultTimeout,
	}

	// TCP dial with context-derived deadline
	start := time.Now()
	dialer := net.Dialer{Timeout: defaultTimeout}
	if deadline, ok := ctx.Deadline(); ok {
		dialer.Deadline = deadline
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		result.LatencyMs = time.Since(start).Milliseconds()
		result.Error = fmt.Sprintf("tcp connection failed: %v", err)
		return result
	}

	result.Reachable = true
	result.LatencyMs = time.Since(start).Milliseconds()

	// SSH handshake
	sshConn, chans, reqs, err := gossh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		result.Error = fmt.Sprintf("ssh handshake failed: %v", err)
		return result
	}

	client := gossh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	result.Authenticated = true
	result.ServerVersion = string(sshConn.ServerVersion())

	return result
}
