package controller

import (
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/store"
)

// Option configures a Controller during construction.
type Option func(*Controller) error

// WithVersion sets the build metadata exposed by the meta provider.
func WithVersion(v meta.VersionInfo) Option {
	return func(c *Controller) error { c.versionInfo = v; return nil }
}

// WithListenAddr sets the address the HTTP server binds to.
func WithListenAddr(addr string) Option {
	return func(c *Controller) error { c.listenAddr = addr; return nil }
}

// WithDebug enables debug-level logging and source locations.
func WithDebug(enabled bool) Option {
	return func(c *Controller) error { c.debug = enabled; return nil }
}

// WithDataDir sets the directory for the SQLite database and auto-generated TLS certs.
func WithDataDir(dir string) Option {
	return func(c *Controller) error { c.dataDir = dir; return nil }
}

// WithTLS configures TLS, mTLS, and auto-cert generation.
func WithTLS(auto bool, cert, key, mtlsCA string) Option {
	return func(c *Controller) error {
		c.tlsAuto = auto
		c.tlsCert = cert
		c.tlsKey = key
		c.tlsMTLSCA = mtlsCA
		return nil
	}
}

// WithAPIToken sets the bearer token for API authentication.
// Falls back to AETHER_API_TOKEN env var during Run() if empty.
func WithAPIToken(token string) Option {
	return func(c *Controller) error { c.apiToken = token; return nil }
}

// WithRBAC enables role-based access control.
func WithRBAC(enabled bool) Option {
	return func(c *Controller) error { c.rbacEnabled = enabled; return nil }
}

// WithFrontend controls embedded/directory frontend serving.
func WithFrontend(enabled bool, dir string) Option {
	return func(c *Controller) error {
		c.frontendEnabled = enabled
		c.frontendDir = dir
		return nil
	}
}

// WithMetrics configures metrics collection interval and retention.
func WithMetrics(interval, retention string) Option {
	return func(c *Controller) error {
		c.metricsInterval = interval
		c.metricsRetention = retention
		return nil
	}
}

// WithEncryptionKey sets the AES-256-GCM key for credential encryption.
func WithEncryptionKey(key string) Option {
	return func(c *Controller) error { c.encryptionKey = key; return nil }
}

// WithStoreOptions appends additional store.Option values passed through to store.New.
func WithStoreOptions(opts ...store.Option) Option {
	return func(c *Controller) error {
		c.storeOpts = append(c.storeOpts, opts...)
		return nil
	}
}

// WithProvider registers a named provider factory to be initialized during Run().
func WithProvider(name string, enabled bool, factory ProviderFactory) Option {
	return func(c *Controller) error {
		c.providerRegs = append(c.providerRegs, providerReg{
			name:    name,
			enabled: enabled,
			factory: factory,
		})
		return nil
	}
}
