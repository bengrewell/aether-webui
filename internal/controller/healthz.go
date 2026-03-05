package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bengrewell/aether-webui/internal/provider"
)

type healthzResponse struct {
	Status  string   `json:"status"`
	Version string   `json:"version"`
	Uptime  string   `json:"uptime"`
	Issues  []string `json:"issues,omitempty"`
}

// registerHealthz adds a lightweight /healthz endpoint directly on the
// transport's router, outside the Huma/OpenAPI spec.
func (c *Controller) registerHealthz(transport Transport, startTime time.Time) {
	transport.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		status := "healthy"
		var issues []string

		type statusInfoer interface {
			StatusInfo() provider.StatusInfo
		}
		for _, p := range c.providers {
			si := p.(statusInfoer).StatusInfo()
			if si.Degraded {
				status = "degraded"
				issues = append(issues, fmt.Sprintf("%s: %s", p.Name(), si.DegradedReason))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(healthzResponse{
			Status:  status,
			Version: c.versionInfo.Version,
			Uptime:  time.Since(startTime).Round(time.Second).String(),
			Issues:  issues,
		})
	})
}
