package provider

import "github.com/bengrewell/aether-webui/internal/endpoint"

type Provider interface {
	Name() string
	Endpoints() []endpoint.AnyEndpoint
	Status() any
	Enable()
	Disable()
	Start() error
	Stop() error
}
