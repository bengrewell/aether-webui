package ssh

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds connection parameters for an SSH session.
type Config struct {
	Host     string        // "host:port"
	User     string
	Password string        // optional
	Key      []byte        // optional PEM-encoded private key
	Timeout  time.Duration // dial timeout; defaults to 10s
}

// Client wraps an SSH connection and provides a simple Run interface.
type Client struct {
	conn *ssh.Client
}

// Dial establishes an SSH connection using the provided config.
// Authentication methods are tried in order: key, then password.
func Dial(ctx context.Context, cfg Config) (*Client, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	var authMethods []ssh.AuthMethod
	if len(cfg.Key) > 0 {
		signer, err := ssh.ParsePrivateKey(cfg.Key)
		if err != nil {
			return nil, fmt.Errorf("ssh: parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("ssh: no authentication method provided")
	}

	sshCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // lab/cluster environment
		Timeout:         timeout,
	}

	host := cfg.Host
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "22")
	}

	// Use a dialer that respects context cancellation.
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, fmt.Errorf("ssh: dial %s: %w", host, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, host, sshCfg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ssh: handshake %s: %w", host, err)
	}

	return &Client{conn: ssh.NewClient(sshConn, chans, reqs)}, nil
}

// Run executes a command on the remote host and returns its stdout, stderr,
// exit code, and any error. The context controls the session lifetime.
func (c *Client) Run(ctx context.Context, cmd string) (stdout, stderr []byte, exitCode int, err error) {
	sess, err := c.conn.NewSession()
	if err != nil {
		return nil, nil, -1, fmt.Errorf("ssh: new session: %w", err)
	}
	defer sess.Close()

	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf

	// Cancel the session when the context is done.
	done := make(chan error, 1)
	go func() {
		done <- sess.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		_ = sess.Signal(ssh.SIGKILL)
		return nil, nil, -1, ctx.Err()
	case runErr := <-done:
		if runErr != nil {
			if exitErr, ok := runErr.(*ssh.ExitError); ok {
				return outBuf.Bytes(), errBuf.Bytes(), exitErr.ExitStatus(), nil
			}
			return outBuf.Bytes(), errBuf.Bytes(), -1, runErr
		}
		return outBuf.Bytes(), errBuf.Bytes(), 0, nil
	}
}

// Close terminates the SSH connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
