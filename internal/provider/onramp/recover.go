package onramp

import (
	"context"
	"log/slog"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
)

// recoverStaleTasks marks any actions and deployments that were "running" or
// "pending" at shutdown as failed. After a service restart, these tasks will
// never complete because the processes that were executing them are gone.
// Calling this on startup prevents the frontend from polling forever on a
// task that will never finish.
func recoverStaleTasks(st store.Client, log *slog.Logger) {
	// Guard against zero-value store (provider started without a backing store).
	if st.Path() == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now().UTC()

	// Recover stale actions (running or pending with no task runner behind them).
	for _, status := range []string{"running", "pending"} {
		actions, err := st.ListActions(ctx, store.ActionFilter{
			Status: status,
			Limit:  1000,
		})
		if err != nil {
			log.Error("failed to list stale actions", "status", status, "error", err)
			continue
		}
		for _, a := range actions {
			result := store.ActionResult{
				Status:     "failed",
				Error:      "service restarted while task was " + status,
				ExitCode:   -1,
				FinishedAt: now,
			}
			if err := st.UpdateActionResult(ctx, a.ID, result); err != nil {
				log.Error("failed to recover stale action", "id", a.ID, "error", err)
			} else {
				log.Warn("recovered stale action", "id", a.ID, "component", a.Component, "action", a.Action, "was", status)
			}
		}
	}

	// Recover stale deployments.
	for _, status := range []string{"running", "pending"} {
		deps, err := st.ListDeployments(ctx, store.DeploymentFilter{
			Status: status,
			Limit:  1000,
		})
		if err != nil {
			log.Error("failed to list stale deployments", "status", status, "error", err)
			continue
		}
		for _, d := range deps {
			if err := st.UpdateDeploymentStatus(ctx, d.ID, "failed", "service restarted while deployment was "+status, now); err != nil {
				log.Error("failed to recover stale deployment", "id", d.ID, "error", err)
			} else {
				log.Warn("recovered stale deployment", "id", d.ID, "was", status)
			}
		}
	}
}
