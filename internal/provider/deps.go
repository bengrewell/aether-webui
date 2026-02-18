package provider

import (
	"log/slog"

	"github.com/bengrewell/aether-webui/internal/store"
)

// WithLogger injects a scoped logger into a provider's Base.
func WithLogger(log *slog.Logger) Option {
	return func(b *Base) { b.log = log }
}

// WithStore injects a store client into a provider's Base.
func WithStore(st store.Client) Option {
	return func(b *Base) { b.store = st }
}

// Log returns the provider's logger, falling back to slog.Default().
func (b *Base) Log() *slog.Logger { return b.log }

// Store returns the provider's store client.
func (b *Base) Store() store.Client { return b.store }
