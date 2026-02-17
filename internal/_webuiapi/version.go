package webuiapi

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

// VersionInfo holds build-time version metadata.
type VersionInfo struct {
	Version    string `json:"version" example:"1.0.0" doc:"Semantic version of the build"`
	BuildDate  string `json:"build_date" example:"2025-01-15T10:30:00Z" doc:"Timestamp when the binary was built"`
	Branch     string `json:"branch" example:"main" doc:"Git branch the binary was built from"`
	CommitHash string `json:"commit_hash" example:"abc1234" doc:"Git commit hash the binary was built from"`
}

// VersionOutput is the response body for the version endpoint.
type VersionOutput struct {
	Body VersionInfo
}

// RegisterVersionRoutes registers the version endpoint with the Huma API.
func RegisterVersionRoutes(api huma.API, info VersionInfo) {
	huma.Register(api, huma.Operation{
		OperationID: "get-version",
		Method:      "GET",
		Path:        "/api/v1/version",
		Summary:     "Build version",
		Description: "Returns build-time version metadata for the running binary",
		Tags:        []string{"Version"},
	}, func(ctx context.Context, input *struct{}) (*VersionOutput, error) {
		return &VersionOutput{Body: info}, nil
	})
}
