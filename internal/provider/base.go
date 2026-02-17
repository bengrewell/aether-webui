package provider

import (
	"sync"

	"github.com/bengrewell/aether-webui/internal/endpoint"
)

type Option func(*Base)

type Base struct {
	name string

	mu      sync.RWMutex
	enabled bool
	running bool
	descs   []endpoint.Descriptor // for status/introspection
	huma    humaHook              // nil if not enabled
	// later: grpcHook, wsHook, etc.
}

func New(name string, opts ...Option) *Base {
	b := &Base{
		mu:      sync.RWMutex{},
		name:    name,
		enabled: true,
		descs:   make([]endpoint.Descriptor, 0, 16),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}
	return b
}

func (b *Base) Name() string { return b.name }

func (b *Base) Enable()  { b.mu.Lock(); b.enabled = true; b.mu.Unlock() }
func (b *Base) Disable() { b.mu.Lock(); b.enabled = false; b.mu.Unlock() }

func (b *Base) Start() error { return nil }
func (b *Base) Stop() error  { return nil }

func (b *Base) addDesc(d endpoint.Descriptor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.descs = append(b.descs, d)
}

func (b *Base) Descriptors() []endpoint.Descriptor {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]endpoint.Descriptor, len(b.descs))
	copy(out, b.descs)
	return out
}

func (b *Base) Status() any {
	type Status struct {
		Enabled bool                  `json:"enabled"`
		Running bool                  `json:"running"`
		Ops     []endpoint.Descriptor `json:"endpoints"`
	}
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]endpoint.Descriptor, len(b.descs))
	copy(out, b.descs)

	return Status{
		Enabled: b.enabled,
		Running: b.running,
		Ops:     out,
	}
}

// StatusInfo is a typed snapshot of a provider's current state for introspection.
type StatusInfo struct {
	Enabled       bool `json:"enabled"`
	Running       bool `json:"running"`
	EndpointCount int  `json:"endpointCount"`
}

// StatusInfo returns a typed snapshot of the provider's enabled/running state and
// endpoint count.
func (b *Base) StatusInfo() StatusInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return StatusInfo{
		Enabled:       b.enabled,
		Running:       b.running,
		EndpointCount: len(b.descs),
	}
}
