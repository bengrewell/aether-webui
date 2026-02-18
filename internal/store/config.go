package store

import "time"

type config struct {
	busyTimeout   time.Duration
	crypter       Crypter
	metricsMaxAge time.Duration
}

func defaults() config {
	return config{
		busyTimeout:   5 * time.Second,
		crypter:       NoopCrypter{},
		metricsMaxAge: 7 * 24 * time.Hour,
	}
}

// Option configures the store returned by New.
type Option func(*config)

// WithBusyTimeout sets the SQLite busy timeout.
func WithBusyTimeout(d time.Duration) Option { return func(c *config) { c.busyTimeout = d } }

// WithCrypter sets the encryption backend for credential secrets.
func WithCrypter(cr Crypter) Option { return func(c *config) { c.crypter = cr } }

// WithMetricsMaxAge sets the retention period for compacting old metric samples.
func WithMetricsMaxAge(d time.Duration) Option { return func(c *config) { c.metricsMaxAge = d } }
