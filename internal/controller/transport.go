package controller

import (
	"net/http"

	"github.com/bengrewell/aether-webui/internal/provider"
)

// Transport abstracts the API layer (e.g. REST, gRPC) that providers
// register endpoints against. rest.Transport satisfies this interface.
type Transport interface {
	ProviderOpts(name string) []provider.Option
	Handler() http.Handler
	Mount(pattern string, h http.Handler)
}
