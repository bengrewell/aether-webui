package onramp

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/bengrewell/aether-webui/internal/store"
	"github.com/bengrewell/aether-webui/internal/taskrunner"
)

// actionCategory classifies an action name as "install", "uninstall", or ""
// (for actions that don't affect component state). Matches install, *-install,
// uninstall, *-uninstall.
func actionCategory(action string) string {
	if action == "install" || strings.HasSuffix(action, "-install") {
		return "install"
	}
	if action == "uninstall" || strings.HasSuffix(action, "-uninstall") {
		return "uninstall"
	}
	return ""
}

// buildOnComplete returns a TaskView callback that persists the action result
// and updates component state when appropriate. The callback is safe to call
// from the task goroutine (no mutex held).
func buildOnComplete(st store.Client, log *slog.Logger, actionID, component, action string) func(taskrunner.TaskView) {
	return func(v taskrunner.TaskView) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		status := string(v.Status)
		result := store.ActionResult{
			Status:     status,
			Error:      v.Error,
			ExitCode:   v.ExitCode,
			FinishedAt: v.FinishedAt,
		}
		if err := st.UpdateActionResult(ctx, actionID, result); err != nil {
			log.Error("failed to update action result", "action_id", actionID, "error", err)
		}

		cat := actionCategory(action)
		if cat == "" {
			return
		}

		var compStatus string
		switch {
		case cat == "install" && v.Status == taskrunner.StatusSucceeded:
			compStatus = "installed"
		case cat == "install":
			compStatus = "failed"
		case cat == "uninstall" && v.Status == taskrunner.StatusSucceeded:
			compStatus = "not_installed"
		case cat == "uninstall":
			compStatus = "failed"
		}

		cs := store.ComponentState{
			Component:  component,
			Status:     compStatus,
			LastAction: action,
			ActionID:   actionID,
			UpdatedAt:  v.FinishedAt,
		}
		if err := st.UpsertComponentState(ctx, cs); err != nil {
			log.Error("failed to update component state", "component", component, "error", err)
		}
	}
}
