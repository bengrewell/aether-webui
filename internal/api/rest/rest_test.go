package rest

import (
	"testing"
)

func TestSecurityScheme_Enabled(t *testing.T) {
	tr := NewTransport(Config{
		APITitle:         "Test API",
		APIVersion:       "0.0.0",
		TokenAuthEnabled: true,
	})

	spec := tr.API().OpenAPI()

	if spec.Components == nil {
		t.Fatal("expected Components to be non-nil")
	}
	scheme, ok := spec.Components.SecuritySchemes["bearerAuth"]
	if !ok {
		t.Fatal("expected bearerAuth security scheme to be defined")
	}
	if scheme.Type != "http" {
		t.Errorf("scheme type = %q, want %q", scheme.Type, "http")
	}
	if scheme.Scheme != "bearer" {
		t.Errorf("scheme scheme = %q, want %q", scheme.Scheme, "bearer")
	}

	if len(spec.Security) == 0 {
		t.Fatal("expected global security requirement to be set")
	}
	if _, ok := spec.Security[0]["bearerAuth"]; !ok {
		t.Error("expected global security to reference bearerAuth")
	}
}

func TestSecurityScheme_Disabled(t *testing.T) {
	tr := NewTransport(Config{
		APITitle:         "Test API",
		APIVersion:       "0.0.0",
		TokenAuthEnabled: false,
	})

	spec := tr.API().OpenAPI()

	if spec.Components != nil && len(spec.Components.SecuritySchemes) > 0 {
		t.Error("expected no security schemes when token auth is disabled")
	}
	if len(spec.Security) > 0 {
		t.Error("expected no global security requirement when token auth is disabled")
	}
}
