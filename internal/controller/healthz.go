package controller

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthzResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

// registerHealthz adds a lightweight /healthz endpoint directly on the
// transport's router, outside the Huma/OpenAPI spec.
func (c *Controller) registerHealthz(transport Transport, startTime time.Time) {
	transport.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(healthzResponse{
			Status:  "healthy",
			Version: c.versionInfo.Version,
			Uptime:  time.Since(startTime).Round(time.Second).String(),
		})
	})
}
