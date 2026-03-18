package onramp

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/bengrewell/aether-webui/internal/store"
	"github.com/bengrewell/aether-webui/internal/taskrunner"
)

// installTier maps component names to their dependency tier for ordering.
// Lower tiers are installed first.
var installTier = map[string]int{
	"k8s":      0,
	"5gc":      1,
	"4gc":      1,
	"amp":      2,
	"sdran":    2,
	"oscric":   2,
	"gnbsim":   3,
	"ueransim": 3,
	"oai":      3,
	"srsran":   3,
	"n3iwf":    3,
	"cluster":  4,
}

// orderActions sorts component/action pairs by dependency tier. Install actions
// sort ascending (dependencies first); all-uninstall sorts descending (dependents
// first). Mixed actions use install order. Stable sort preserves user-specified
// sub-order within the same tier.
func orderActions(pairs []ComponentActionPair) []ComponentActionPair {
	out := make([]ComponentActionPair, len(pairs))
	copy(out, pairs)

	allUninstall := true
	for _, p := range out {
		if actionCategory(p.Action) != "uninstall" {
			allUninstall = false
			break
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		ti := installTier[out[i].Component]
		tj := installTier[out[j].Component]
		if allUninstall {
			return ti > tj // reverse order for uninstall
		}
		return ti < tj
	})

	return out
}

// HandleDeploy validates and submits a batch deployment.
func (o *OnRamp) HandleDeploy(ctx context.Context, in *DeployInput) (*DeployOutput, error) {
	if len(in.Body.Actions) == 0 {
		return nil, huma.Error422UnprocessableEntity("actions list must not be empty")
	}

	// Validate all component/action pairs.
	for _, pair := range in.Body.Actions {
		comp, ok := componentIndex[pair.Component]
		if !ok {
			return nil, huma.Error422UnprocessableEntity(
				fmt.Sprintf("unknown component: %s", pair.Component))
		}
		found := false
		for _, a := range comp.Actions {
			if a.Name == pair.Action {
				found = true
				break
			}
		}
		if !found {
			return nil, huma.Error422UnprocessableEntity(
				fmt.Sprintf("component %s has no action %s", pair.Component, pair.Action))
		}
	}

	ordered := orderActions(in.Body.Actions)

	deployID := uuid.NewString()
	now := time.Now().UTC()
	st := o.Store()
	log := o.Log()

	dep := store.Deployment{
		ID:        deployID,
		Status:    "running",
		CreatedAt: now,
		StartedAt: now,
	}

	// Build deployment actions with pre-generated IDs.
	for i, pair := range ordered {
		actionID := uuid.NewString()
		dep.Actions = append(dep.Actions, store.DeploymentAction{
			DeploymentID: deployID,
			Seq:          i,
			ActionID:     actionID,
			Component:    pair.Component,
			Action:       pair.Action,
		})
	}

	// Insert deployment record first so action_history rows are never orphaned.
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	if err := st.InsertDeployment(dbCtx, dep); err != nil {
		log.Error("failed to insert deployment", "deployment_id", deployID, "error", err)
		return nil, huma.Error500InternalServerError("failed to create deployment", err)
	}

	// Insert action_history records for each action.
	for _, da := range dep.Actions {
		target := resolveTarget(da.Component, da.Action)
		rec := store.ActionRecord{
			ID:        da.ActionID,
			Component: da.Component,
			Action:    da.Action,
			Target:    target,
			Status:    "pending",
			ExitCode:  -1,
			StartedAt: now,
		}
		if err := st.InsertAction(dbCtx, rec); err != nil {
			log.Error("failed to insert action record for deployment", "action_id", da.ActionID, "error", err)
			_ = st.UpdateDeploymentStatus(dbCtx, deployID, "failed", err.Error(), time.Now().UTC())
			o.cancelRemainingActions(dep, 0)
			return nil, huma.Error500InternalServerError("failed to create deployment", err)
		}
	}

	// Submit the first action. The deployment is already "running" so there is
	// no race if the task completes before this function returns.
	first := dep.Actions[0]
	target := resolveTarget(first.Component, first.Action)

	if err := o.submitDeploymentAction(dep, 0, first.ActionID, first.Component, first.Action, target); err != nil {
		_ = st.UpdateDeploymentStatus(dbCtx, deployID, "failed", err.Error(), time.Now().UTC())
		o.cancelRemainingActions(dep, 0)
		return nil, huma.Error500InternalServerError("failed to start deployment", err)
	}

	return &DeployOutput{Body: o.buildDeploymentItem(ctx, dep)}, nil
}

// submitDeploymentAction submits one action from a deployment to the task runner
// with chained OnComplete logic.
func (o *OnRamp) submitDeploymentAction(dep store.Deployment, seq int, actionID, component, action, target string) error {
	st := o.Store()
	log := o.Log()

	baseOnComplete := buildOnComplete(st, log, actionID, component, action)
	baseOnStart := buildOnStart(st, log, actionID, component, action)

	chainedOnComplete := func(v taskrunner.TaskView) {
		// Run the standard action_history + component_state updates.
		baseOnComplete(v)

		dCtx, dCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer dCancel()

		// Check if the deployment was canceled while this action ran.
		currentDep, found, err := st.GetDeployment(dCtx, dep.ID)
		if err != nil || !found {
			log.Error("failed to load deployment in OnComplete", "deployment_id", dep.ID, "error", err)
			return
		}
		if currentDep.Status == "canceled" {
			return
		}

		isLast := seq == len(dep.Actions)-1

		switch v.Status {
		case taskrunner.StatusSucceeded:
			if isLast {
				if err := st.UpdateDeploymentStatus(dCtx, dep.ID, "succeeded", "", time.Now().UTC()); err != nil {
					log.Error("failed to mark deployment succeeded", "deployment_id", dep.ID, "error", err)
				}
			} else {
				// Submit the next action.
				next := dep.Actions[seq+1]
				nextTarget := resolveTarget(next.Component, next.Action)
				if err := o.submitDeploymentAction(dep, seq+1, next.ActionID, next.Component, next.Action, nextTarget); err != nil {
					log.Error("failed to submit next deployment action", "deployment_id", dep.ID, "seq", seq+1, "error", err)
					_ = st.UpdateDeploymentStatus(dCtx, dep.ID, "failed", err.Error(), time.Now().UTC())
					o.cancelRemainingActions(dep, seq+1)
				}
			}

		case taskrunner.StatusFailed, taskrunner.StatusCanceled:
			errMsg := v.Error
			if errMsg == "" {
				errMsg = fmt.Sprintf("action %s/%s failed", component, action)
			}
			if err := st.UpdateDeploymentStatus(dCtx, dep.ID, "failed", errMsg, time.Now().UTC()); err != nil {
				log.Error("failed to mark deployment failed", "deployment_id", dep.ID, "error", err)
			}
			// Cancel remaining actions.
			o.cancelRemainingActions(dep, seq+1)
		}
	}

	_, err := o.runner.Submit(taskrunner.TaskSpec{
		ID:          actionID,
		Command:     "make",
		Args:        []string{target},
		Dir:         o.config.OnRampDir,
		Description: fmt.Sprintf("deploy:%s/%s", component, action),
		Labels: map[string]string{
			"component":     component,
			"action":        action,
			"target":        target,
			"deployment_id": dep.ID,
		},
		OnStart:    baseOnStart,
		OnComplete: chainedOnComplete,
	})
	return err
}

// cancelRemainingActions marks all actions from startSeq onward as canceled.
func (o *OnRamp) cancelRemainingActions(dep store.Deployment, startSeq int) {
	st := o.Store()
	log := o.Log()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := startSeq; i < len(dep.Actions); i++ {
		a := dep.Actions[i]
		result := store.ActionResult{
			Status:     "canceled",
			Error:      "deployment failed or canceled",
			ExitCode:   -1,
			FinishedAt: time.Now().UTC(),
		}
		if err := st.UpdateActionResult(ctx, a.ActionID, result); err != nil {
			log.Error("failed to cancel remaining action", "action_id", a.ActionID, "error", err)
		}
	}
}

// HandleGetDeployment returns a single deployment with enriched action statuses.
func (o *OnRamp) HandleGetDeployment(ctx context.Context, in *DeploymentGetInput) (*DeploymentGetOutput, error) {
	dep, found, err := o.Store().GetDeployment(ctx, in.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get deployment", err)
	}
	if !found {
		return nil, huma.Error404NotFound("deployment not found", fmt.Errorf("no deployment with id %s", in.ID))
	}
	return &DeploymentGetOutput{Body: o.buildDeploymentItem(ctx, dep)}, nil
}

// HandleListDeployments returns a paginated list of deployments.
func (o *OnRamp) HandleListDeployments(ctx context.Context, in *DeploymentListInput) (*DeploymentListOutput, error) {
	deps, err := o.Store().ListDeployments(ctx, store.DeploymentFilter{
		Status: in.Status,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list deployments", err)
	}

	items := make([]DeploymentItem, len(deps))
	for i, dep := range deps {
		items[i] = o.buildDeploymentItem(ctx, dep)
	}
	return &DeploymentListOutput{Body: items}, nil
}

// HandleCancelDeployment cancels a running or pending deployment.
func (o *OnRamp) HandleCancelDeployment(ctx context.Context, in *DeploymentCancelInput) (*DeploymentCancelOutput, error) {
	st := o.Store()

	dep, found, err := st.GetDeployment(ctx, in.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get deployment", err)
	}
	if !found {
		return nil, huma.Error404NotFound("deployment not found", fmt.Errorf("no deployment with id %s", in.ID))
	}

	// Reject if already terminal.
	switch dep.Status {
	case "succeeded", "failed", "canceled":
		return nil, huma.Error409Conflict(
			fmt.Sprintf("deployment already %s", dep.Status))
	}

	// Mark deployment as canceled.
	if err := st.UpdateDeploymentStatus(ctx, dep.ID, "canceled", "canceled by user", time.Now().UTC()); err != nil {
		return nil, huma.Error500InternalServerError("failed to cancel deployment", err)
	}

	// Attempt to cancel every action in the runner unconditionally to handle
	// the race where a task is submitted between the status check and cancel.
	// ErrNotFound/ErrNotRunning are expected for tasks that haven't started or
	// have already finished.
	for _, a := range dep.Actions {
		_ = o.runner.Cancel(a.ActionID)

		rec, ok, err := st.GetAction(ctx, a.ActionID)
		if err != nil || !ok {
			continue
		}
		if rec.Status == "pending" || rec.Status == "running" {
			result := store.ActionResult{
				Status:     "canceled",
				Error:      "deployment canceled",
				ExitCode:   -1,
				FinishedAt: time.Now().UTC(),
			}
			_ = st.UpdateActionResult(ctx, a.ActionID, result)
		}
	}

	out := &DeploymentCancelOutput{}
	out.Body.Message = fmt.Sprintf("deployment %s canceled", in.ID)
	return out, nil
}

// buildDeploymentItem converts a store.Deployment to the API response type,
// enriching each action's status from action_history.
func (o *OnRamp) buildDeploymentItem(ctx context.Context, dep store.Deployment) DeploymentItem {
	st := o.Store()

	item := DeploymentItem{
		ID:      dep.ID,
		Status:  dep.Status,
		Actions: make([]DeploymentActionItem, len(dep.Actions)),
	}
	if !dep.CreatedAt.IsZero() {
		item.CreatedAt = dep.CreatedAt.Unix()
	}
	if !dep.StartedAt.IsZero() {
		item.StartedAt = dep.StartedAt.Unix()
	}
	if !dep.FinishedAt.IsZero() {
		item.FinishedAt = dep.FinishedAt.Unix()
	}
	item.Error = dep.Error

	for i, a := range dep.Actions {
		dai := DeploymentActionItem{
			Seq:       a.Seq,
			ActionID:  a.ActionID,
			Component: a.Component,
			Action:    a.Action,
			Status:    "pending",
		}
		if rec, ok, err := st.GetAction(ctx, a.ActionID); err == nil && ok {
			dai.Status = rec.Status
		}
		item.Actions[i] = dai
	}

	return item
}

// resolveTarget looks up the Makefile target for a component/action pair.
func resolveTarget(component, action string) string {
	comp, ok := componentIndex[component]
	if !ok {
		return ""
	}
	for _, a := range comp.Actions {
		if a.Name == action {
			return a.Target
		}
	}
	return ""
}

