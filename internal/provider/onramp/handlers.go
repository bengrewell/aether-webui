package onramp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/bengrewell/aether-webui/internal/store"
	"github.com/bengrewell/aether-webui/internal/taskrunner"
)

// ---------------------------------------------------------------------------
// Repo handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleGetRepoStatus(_ context.Context, _ *struct{}) (*RepoStatusOutput, error) {
	status := o.gatherRepoStatus()
	return &RepoStatusOutput{Body: status}, nil
}

func (o *OnRamp) handleRefreshRepo(_ context.Context, _ *struct{}) (*RepoRefreshOutput, error) {
	log := o.Log()
	if err := ensureRepo(o.config, log); err != nil {
		status := o.gatherRepoStatus()
		status.Error = err.Error()
		return &RepoRefreshOutput{Body: status}, nil
	}
	status := o.gatherRepoStatus()
	return &RepoRefreshOutput{Body: status}, nil
}

// gatherRepoStatus inspects the OnRamp directory and returns its git state.
func (o *OnRamp) gatherRepoStatus() RepoStatus {
	rs := RepoStatus{
		Dir:     o.config.OnRampDir,
		RepoURL: o.config.RepoURL,
		Version: o.config.Version,
	}

	gitDir := filepath.Join(o.config.OnRampDir, ".git")
	if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
		return rs
	}
	rs.Cloned = true

	if commit, err := gitOutput(o.config.OnRampDir, "rev-parse", "HEAD"); err == nil {
		rs.Commit = commit
	}

	if branch, err := gitOutput(o.config.OnRampDir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		rs.Branch = branch
	}

	// Resolve the tag pointing at HEAD, if any.
	if tag, err := gitOutput(o.config.OnRampDir, "describe", "--tags", "--exact-match", "HEAD"); err == nil {
		rs.Tag = tag
	}

	// A non-empty output from `git status --porcelain` indicates uncommitted changes.
	if porcelain, err := gitOutput(o.config.OnRampDir, "status", "--porcelain"); err == nil && porcelain != "" {
		rs.Dirty = true
	}

	return rs
}

// ---------------------------------------------------------------------------
// Component handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleListComponents(_ context.Context, _ *struct{}) (*ComponentListOutput, error) {
	return &ComponentListOutput{Body: componentRegistry}, nil
}

func (o *OnRamp) handleGetComponent(_ context.Context, in *ComponentGetInput) (*ComponentGetOutput, error) {
	comp, ok := componentIndex[in.Component]
	if !ok {
		return nil, huma.Error404NotFound("component not found", fmt.Errorf("unknown component: %s", in.Component))
	}
	return &ComponentGetOutput{Body: *comp}, nil
}

func (o *OnRamp) handleExecuteAction(ctx context.Context, in *ExecuteActionInput) (*ExecuteActionOutput, error) {
	comp, ok := componentIndex[in.Component]
	if !ok {
		return nil, huma.Error404NotFound("component not found", fmt.Errorf("unknown component: %s", in.Component))
	}

	var target string
	for _, a := range comp.Actions {
		if a.Name == in.Action {
			target = a.Target
			break
		}
	}
	if target == "" {
		return nil, huma.Error404NotFound("action not found",
			fmt.Errorf("component %s has no action %s", in.Component, in.Action))
	}

	// Extract optional labels/tags from the request body.
	var labels map[string]string
	var tags []string
	if in.Body != nil {
		labels = in.Body.Labels
		tags = in.Body.Tags
	}

	actionID := uuid.NewString()
	st := o.Store()
	log := o.Log()

	view, err := o.runner.Submit(taskrunner.TaskSpec{
		Command:     "make",
		Args:        []string{target},
		Dir:         o.config.OnRampDir,
		Description: fmt.Sprintf("%s/%s", in.Component, in.Action),
		Labels: map[string]string{
			"component": in.Component,
			"action":    in.Action,
			"target":    target,
		},
		OnComplete: buildOnComplete(st, log, actionID, in.Component, in.Action),
	})
	if err != nil {
		if errors.Is(err, taskrunner.ErrConcurrencyLimit) {
			return nil, huma.Error409Conflict("a task is already running", err)
		}
		return nil, huma.Error500InternalServerError("failed to start task", err)
	}

	// Record the action in persistent history.
	now := time.Now().UTC()
	rec := store.ActionRecord{
		ID:        actionID,
		Component: in.Component,
		Action:    in.Action,
		Target:    target,
		Status:    "running",
		ExitCode:  -1,
		Labels:    labels,
		Tags:      tags,
		StartedAt: now,
	}
	if err := st.InsertAction(ctx, rec); err != nil {
		log.Error("failed to insert action record", "action_id", actionID, "error", err)
	}

	// Update component state for install/uninstall actions.
	if cat := actionCategory(in.Action); cat != "" {
		status := "installing"
		if cat == "uninstall" {
			status = "uninstalling"
		}
		cs := store.ComponentState{
			Component:  in.Component,
			Status:     status,
			LastAction: in.Action,
			ActionID:   actionID,
			UpdatedAt:  now,
		}
		if err := st.UpsertComponentState(ctx, cs); err != nil {
			log.Error("failed to upsert component state", "component", in.Component, "error", err)
		}
	}

	return &ExecuteActionOutput{Body: toOnRampTask(view, "", 0)}, nil
}

// ---------------------------------------------------------------------------
// Task handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleListTasks(_ context.Context, _ *struct{}) (*TaskListOutput, error) {
	views := o.runner.List(nil)
	out := make([]OnRampTask, len(views))
	for i, v := range views {
		chunk, _ := o.runner.Output(v.ID, 0)
		out[i] = toOnRampTask(v, chunk.Data, chunk.NewOffset)
	}
	return &TaskListOutput{Body: out}, nil
}

func (o *OnRamp) handleGetTask(_ context.Context, in *TaskGetInput) (*TaskGetOutput, error) {
	view, err := o.runner.Get(in.ID)
	if err != nil {
		return nil, huma.Error404NotFound("task not found", fmt.Errorf("no task with id %s", in.ID))
	}
	chunk, _ := o.runner.Output(in.ID, in.Offset)
	return &TaskGetOutput{Body: toOnRampTask(view, chunk.Data, chunk.NewOffset)}, nil
}

// ---------------------------------------------------------------------------
// Action history handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleListActions(ctx context.Context, in *ActionListInput) (*ActionListOutput, error) {
	recs, err := o.Store().ListActions(ctx, store.ActionFilter{
		Component: in.Component,
		Action:    in.Action,
		Status:    in.Status,
		Limit:     in.Limit,
		Offset:    in.Offset,
	})
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list actions", err)
	}
	items := make([]ActionHistoryItem, len(recs))
	for i, r := range recs {
		items[i] = actionRecordToItem(r)
	}
	return &ActionListOutput{Body: items}, nil
}

func (o *OnRamp) handleGetAction(ctx context.Context, in *ActionGetInput) (*ActionGetOutput, error) {
	rec, ok, err := o.Store().GetAction(ctx, in.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get action", err)
	}
	if !ok {
		return nil, huma.Error404NotFound("action not found", fmt.Errorf("no action with id %s", in.ID))
	}
	return &ActionGetOutput{Body: actionRecordToItem(rec)}, nil
}

func actionRecordToItem(r store.ActionRecord) ActionHistoryItem {
	item := ActionHistoryItem{
		ID:        r.ID,
		Component: r.Component,
		Action:    r.Action,
		Target:    r.Target,
		Status:    r.Status,
		ExitCode:  r.ExitCode,
		Error:     r.Error,
		Labels:    r.Labels,
		Tags:      r.Tags,
		StartedAt: r.StartedAt.Unix(),
	}
	if !r.FinishedAt.IsZero() {
		item.FinishedAt = r.FinishedAt.Unix()
	}
	return item
}

// ---------------------------------------------------------------------------
// Component state handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleListComponentStates(ctx context.Context, _ *struct{}) (*ComponentStateListOutput, error) {
	states, err := o.Store().ListComponentStates(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list component states", err)
	}

	// Build a map of known states.
	stateMap := make(map[string]store.ComponentState, len(states))
	for _, s := range states {
		stateMap[s.Component] = s
	}

	// Emit an entry for every registered component, defaulting to not_installed.
	items := make([]ComponentStateItem, 0, len(componentRegistry))
	for _, comp := range componentRegistry {
		if s, ok := stateMap[comp.Name]; ok {
			items = append(items, componentStateToItem(s))
		} else {
			items = append(items, ComponentStateItem{
				Component: comp.Name,
				Status:    "not_installed",
			})
		}
	}
	return &ComponentStateListOutput{Body: items}, nil
}

func (o *OnRamp) handleGetComponentState(ctx context.Context, in *ComponentStateGetInput) (*ComponentStateGetOutput, error) {
	// Validate the component name against the registry.
	if _, ok := componentIndex[in.Component]; !ok {
		return nil, huma.Error404NotFound("component not found", fmt.Errorf("unknown component: %s", in.Component))
	}

	cs, ok, err := o.Store().GetComponentState(ctx, in.Component)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get component state", err)
	}
	if !ok {
		return &ComponentStateGetOutput{Body: ComponentStateItem{
			Component: in.Component,
			Status:    "not_installed",
		}}, nil
	}
	return &ComponentStateGetOutput{Body: componentStateToItem(cs)}, nil
}

func componentStateToItem(cs store.ComponentState) ComponentStateItem {
	item := ComponentStateItem{
		Component:  cs.Component,
		Status:     cs.Status,
		LastAction: cs.LastAction,
		ActionID:   cs.ActionID,
	}
	if !cs.UpdatedAt.IsZero() {
		item.UpdatedAt = cs.UpdatedAt.Unix()
	}
	return item
}

// ---------------------------------------------------------------------------
// Config handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleGetConfig(_ context.Context, _ *struct{}) (*ConfigGetOutput, error) {
	cfg, err := o.readVarsFile(filepath.Join(o.config.OnRampDir, "vars", "main.yml"))
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read config", err)
	}
	return &ConfigGetOutput{Body: cfg}, nil
}

func (o *OnRamp) handlePatchConfig(_ context.Context, in *ConfigPatchInput) (*ConfigPatchOutput, error) {
	mainYML := filepath.Join(o.config.OnRampDir, "vars", "main.yml")

	base, err := o.readVarsFile(mainYML)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read current config", err)
	}

	mergeConfig(&base, &in.Body)

	if err := o.writeVarsFile(mainYML, &base); err != nil {
		return nil, huma.Error500InternalServerError("failed to write config", err)
	}
	return &ConfigPatchOutput{Body: base}, nil
}

// ---------------------------------------------------------------------------
// Profile handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleListProfiles(_ context.Context, _ *struct{}) (*ProfileListOutput, error) {
	pattern := filepath.Join(o.config.OnRampDir, "vars", "main-*.yml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list profiles", err)
	}
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		base := filepath.Base(m)
		name := strings.TrimPrefix(base, "main-")
		name = strings.TrimSuffix(name, ".yml")
		names = append(names, name)
	}
	return &ProfileListOutput{Body: names}, nil
}

func (o *OnRamp) handleGetProfile(_ context.Context, in *ProfileGetInput) (*ProfileGetOutput, error) {
	path := filepath.Join(o.config.OnRampDir, "vars", fmt.Sprintf("main-%s.yml", in.Name))
	cfg, err := o.readVarsFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, huma.Error404NotFound("profile not found",
				fmt.Errorf("no profile named %s", in.Name))
		}
		return nil, huma.Error500InternalServerError("failed to read profile", err)
	}
	return &ProfileGetOutput{Body: cfg}, nil
}

func (o *OnRamp) handleActivateProfile(_ context.Context, in *ProfileActivateInput) (*ProfileActivateOutput, error) {
	src := filepath.Join(o.config.OnRampDir, "vars", fmt.Sprintf("main-%s.yml", in.Name))
	dst := filepath.Join(o.config.OnRampDir, "vars", "main.yml")

	srcFile, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, huma.Error404NotFound("profile not found",
				fmt.Errorf("no profile named %s", in.Name))
		}
		return nil, huma.Error500InternalServerError("failed to open profile", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to write main.yml", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return nil, huma.Error500InternalServerError("failed to copy profile", err)
	}

	out := &ProfileActivateOutput{}
	out.Body.Message = fmt.Sprintf("profile %q activated", in.Name)
	return out, nil
}

// ---------------------------------------------------------------------------
// YAML helpers
// ---------------------------------------------------------------------------

func (o *OnRamp) readVarsFile(path string) (OnRampConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return OnRampConfig{}, err
	}
	var cfg OnRampConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return OnRampConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

func (o *OnRamp) writeVarsFile(path string, cfg *OnRampConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// ---------------------------------------------------------------------------
// Config merge
// ---------------------------------------------------------------------------

// mergeConfig overwrites non-nil fields in base with values from patch.
func mergeConfig(base, patch *OnRampConfig) {
	if patch.K8s != nil {
		base.K8s = patch.K8s
	}
	if patch.Core != nil {
		base.Core = patch.Core
	}
	if patch.GNBSim != nil {
		base.GNBSim = patch.GNBSim
	}
	if patch.AMP != nil {
		base.AMP = patch.AMP
	}
	if patch.SDRAN != nil {
		base.SDRAN = patch.SDRAN
	}
	if patch.UERANSIM != nil {
		base.UERANSIM = patch.UERANSIM
	}
	if patch.OAI != nil {
		base.OAI = patch.OAI
	}
	if patch.SRSRan != nil {
		base.SRSRan = patch.SRSRan
	}
	if patch.N3IWF != nil {
		base.N3IWF = patch.N3IWF
	}
}
