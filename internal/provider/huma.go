package provider

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/bengrewell/aether-webui/internal/endpoint"
)

type humaHook struct {
	api huma.API
}

func WithHuma(api huma.API) Option {
	return func(b *Base) {
		if api == nil {
			return
		}
		b.huma = humaHook{api: api}
	}
}

func methodFor(s endpoint.Semantics) string {
	switch s {
	case endpoint.Read:
		return http.MethodGet
	case endpoint.Create:
		return http.MethodPost
	case endpoint.Update:
		return http.MethodPut
	case endpoint.Delete:
		return http.MethodDelete
	default:
		return http.MethodPost
	}
}

func normalizePath(p string) string {
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	return p
}

func opFrom(d endpoint.Descriptor) huma.Operation {
	m := d.HTTP.Method
	if m == "" {
		m = methodFor(d.Semantics)
	}
	p := d.HTTP.Path
	if p == "" {
		p = "/api/v1/" + d.OperationID
	}
	p = normalizePath(p)

	return huma.Operation{
		OperationID: d.OperationID,
		Method:      m,
		Path:        p,
		Summary:     d.Summary,
		Description: d.Description,
		Tags:        d.Tags,
	}
}

// Register registers a typed endpoint with any enabled transports (Huma for now).
// If Huma isn't enabled, it just records the descriptor for Status().
func Register[I any, O any](b *Base, ep endpoint.Endpoint[I, O]) {
	b.addDesc(ep.Desc)

	if b.huma.api != nil {
		huma.Register[I, O](b.huma.api, opFrom(ep.Desc), ep.Handler)
	}
}
