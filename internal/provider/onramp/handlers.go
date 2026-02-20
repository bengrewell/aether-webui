package onramp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
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

func (o *OnRamp) handleExecuteAction(_ context.Context, in *ExecuteActionInput) (*ExecuteActionOutput, error) {
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

	// Reject if another task is still running.
	o.mu.Lock()
	for _, t := range o.tasks {
		if t.Status == "running" {
			o.mu.Unlock()
			return nil, huma.Error409Conflict("a task is already running",
				fmt.Errorf("task %s (%s) is still in progress", t.ID, t.Target))
		}
	}

	task := &Task{
		ID:        uuid.NewString(),
		Component: in.Component,
		Action:    in.Action,
		Target:    target,
		Status:    "running",
		StartedAt: time.Now().UTC(),
	}
	// Prepend so most-recent is first.
	o.tasks = append([]*Task{task}, o.tasks...)
	o.mu.Unlock()

	go o.runMake(task)

	return &ExecuteActionOutput{Body: *task}, nil
}

// runMake executes `make <target>` in the OnRamp directory and updates the task
// when complete.
func (o *OnRamp) runMake(task *Task) {
	log := o.Log()

	cmd := exec.Command("make", task.Target)
	cmd.Dir = o.config.OnRampDir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	log.Info("starting make target", "target", task.Target, "task_id", task.ID)
	err := cmd.Run()

	o.mu.Lock()
	task.FinishedAt = time.Now().UTC()
	task.Output = buf.String()
	if err != nil {
		task.Status = "failed"
		if exitErr, ok := err.(*exec.ExitError); ok {
			task.ExitCode = exitErr.ExitCode()
		} else {
			task.ExitCode = -1
		}
		log.Error("make target failed", "target", task.Target, "task_id", task.ID, "error", err)
	} else {
		task.Status = "succeeded"
		task.ExitCode = 0
		log.Info("make target succeeded", "target", task.Target, "task_id", task.ID)
	}
	o.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Task handlers
// ---------------------------------------------------------------------------

func (o *OnRamp) handleListTasks(_ context.Context, _ *struct{}) (*TaskListOutput, error) {
	o.mu.Lock()
	out := make([]Task, len(o.tasks))
	for i, t := range o.tasks {
		out[i] = *t
	}
	o.mu.Unlock()
	return &TaskListOutput{Body: out}, nil
}

func (o *OnRamp) handleGetTask(_ context.Context, in *TaskGetInput) (*TaskGetOutput, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, t := range o.tasks {
		if t.ID == in.ID {
			return &TaskGetOutput{Body: *t}, nil
		}
	}
	return nil, huma.Error404NotFound("task not found", fmt.Errorf("no task with id %s", in.ID))
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
