// Command generate-openapi produces a static OpenAPI 3.1 JSON spec from the
// registered providers. It creates a REST transport, registers every provider
// (no HTTP server or store required), and writes the spec to api/openapi.json.
//
// Usage:
//
//	go run ./cmd/generate-openapi
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/bengrewell/aether-webui/internal/api/rest"
	"github.com/bengrewell/aether-webui/internal/provider/meta"
	"github.com/bengrewell/aether-webui/internal/provider/nodes"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/provider/system"
)

func main() {
	transport := rest.NewTransport(rest.Config{
		APITitle:         "Aether WebUI API",
		APIVersion:       "0.0.0",
		TokenAuthEnabled: true,
	})

	// Register all providers. Only Huma API registration matters — no store or
	// runtime functionality is needed.
	opts := transport.ProviderOpts

	system.NewProvider(system.Config{CollectInterval: 10 * time.Second}, opts("system")...)
	nodes.NewProvider(opts("nodes")...)
	onramp.NewProvider(onramp.Config{}, opts("onramp")...)
	meta.NewProvider(
		meta.VersionInfo{},
		meta.AppConfig{},
		nil, nil, nil,
		opts("meta")...,
	)

	spec := transport.API().OpenAPI()
	b, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling spec: %v\n", err)
		os.Exit(1)
	}

	// Append trailing newline for POSIX compliance.
	b = append(b, '\n')

	if err := os.WriteFile("api/openapi.json", b, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("wrote api/openapi.json")
}
