package webuiapi

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

// HealthOutput is the response body for the health check endpoint.
type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"healthy" doc:"Health status of the service"`
	}
}

// RegisterHealthRoutes registers health-related routes with the Huma API.
func RegisterHealthRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-health",
		Method:      "GET",
		Path:        "/healthz",
		Summary:     "Health check",
		Description: "Returns the health status of the service",
		Tags:        []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		resp := &HealthOutput{}
		resp.Body.Status = "healthy"
		return resp, nil
	})
}
