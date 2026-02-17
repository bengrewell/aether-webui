package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// RunHelmInstall installs a Helm chart.
func (e *DefaultExecutor) RunHelmInstall(ctx context.Context, opts HelmInstallOptions) (*ExecResult, error) {
	args := []string{"install", opts.ReleaseName, opts.Chart}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	if opts.Version != "" {
		args = append(args, "--version", opts.Version)
	}
	if opts.Repo != "" {
		args = append(args, "--repo", opts.Repo)
	}
	for _, f := range opts.ValuesFiles {
		args = append(args, "--values", f)
	}
	for k, v := range opts.Values {
		args = append(args, "--set", fmt.Sprintf("%s=%v", k, v))
	}
	if opts.Wait {
		args = append(args, "--wait")
		if opts.WaitTimeout > 0 {
			args = append(args, "--timeout", opts.WaitTimeout.String())
		}
	}
	if opts.CreateNS {
		args = append(args, "--create-namespace")
	}
	if opts.Atomic {
		args = append(args, "--atomic")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	if opts.Description != "" {
		args = append(args, "--description", opts.Description)
	}
	if opts.Replace {
		args = append(args, "--replace")
	}

	return e.runCommand(ctx, opts.BaseOptions, "helm", args...)
}

// RunHelmUpgrade upgrades a Helm release.
func (e *DefaultExecutor) RunHelmUpgrade(ctx context.Context, opts HelmUpgradeOptions) (*ExecResult, error) {
	args := []string{"upgrade", opts.ReleaseName, opts.Chart}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	if opts.Version != "" {
		args = append(args, "--version", opts.Version)
	}
	if opts.Repo != "" {
		args = append(args, "--repo", opts.Repo)
	}
	for _, f := range opts.ValuesFiles {
		args = append(args, "--values", f)
	}
	for k, v := range opts.Values {
		args = append(args, "--set", fmt.Sprintf("%s=%v", k, v))
	}
	if opts.Wait {
		args = append(args, "--wait")
		if opts.WaitTimeout > 0 {
			args = append(args, "--timeout", opts.WaitTimeout.String())
		}
	}
	if opts.Install {
		args = append(args, "--install")
	}
	if opts.Atomic {
		args = append(args, "--atomic")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	if opts.Description != "" {
		args = append(args, "--description", opts.Description)
	}
	if opts.ReuseValues {
		args = append(args, "--reuse-values")
	}
	if opts.ResetValues {
		args = append(args, "--reset-values")
	}
	if opts.Force {
		args = append(args, "--force")
	}
	if opts.CleanupOnFail {
		args = append(args, "--cleanup-on-fail")
	}

	return e.runCommand(ctx, opts.BaseOptions, "helm", args...)
}

// RunHelmUninstall uninstalls a Helm release.
func (e *DefaultExecutor) RunHelmUninstall(ctx context.Context, opts HelmUninstallOptions) (*ExecResult, error) {
	args := []string{"uninstall", opts.ReleaseName}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	if opts.KeepHistory {
		args = append(args, "--keep-history")
	}
	if opts.DryRun {
		args = append(args, "--dry-run")
	}
	if opts.Wait {
		args = append(args, "--wait")
		if opts.WaitTimeout > 0 {
			args = append(args, "--timeout", opts.WaitTimeout.String())
		}
	}
	if opts.Description != "" {
		args = append(args, "--description", opts.Description)
	}

	return e.runCommand(ctx, opts.BaseOptions, "helm", args...)
}

// RunHelmList lists Helm releases.
func (e *DefaultExecutor) RunHelmList(ctx context.Context, opts HelmListOptions) (*HelmReleaseList, error) {
	args := []string{"list", "--output", "json"}

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	if opts.AllNS {
		args = append(args, "--all-namespaces")
	}
	if opts.Filter != "" {
		args = append(args, "--filter", opts.Filter)
	}
	if opts.All {
		args = append(args, "--all")
	} else {
		if opts.Deployed {
			args = append(args, "--deployed")
		}
		if opts.Failed {
			args = append(args, "--failed")
		}
		if opts.Pending {
			args = append(args, "--pending")
		}
		if opts.Superseded {
			args = append(args, "--superseded")
		}
		if opts.Uninstalled {
			args = append(args, "--uninstalled")
		}
	}
	if opts.Offset > 0 {
		args = append(args, "--offset", strconv.Itoa(opts.Offset))
	}
	if opts.Max > 0 {
		args = append(args, "--max", strconv.Itoa(opts.Max))
	}
	if opts.SortBy != "" {
		args = append(args, "--date") // helm list only has --date flag
		if opts.SortBy == "date" {
			args = append(args, "--date")
		}
	}
	if opts.Reverse {
		args = append(args, "--reverse")
	}

	result, err := e.runCommand(ctx, opts.BaseOptions, "helm", args...)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("helm list failed: %s", result.Stderr)
	}

	// Parse JSON output
	var releases []helmListEntry
	if err := json.Unmarshal([]byte(result.Stdout), &releases); err != nil {
		return nil, fmt.Errorf("failed to parse helm list output: %w", err)
	}

	list := &HelmReleaseList{
		Releases: make([]HelmRelease, 0, len(releases)),
	}

	for _, r := range releases {
		updated, _ := time.Parse(time.RFC3339, r.Updated)
		list.Releases = append(list.Releases, HelmRelease{
			Name:       r.Name,
			Namespace:  r.Namespace,
			Revision:   parseRevision(r.Revision),
			Status:     r.Status,
			Chart:      r.Chart,
			AppVersion: r.AppVersion,
			Updated:    updated,
		})
	}

	return list, nil
}

// helmListEntry represents the JSON output from helm list.
type helmListEntry struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

// parseRevision converts a revision string to int.
func parseRevision(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// RunHelmStatus gets the status of a Helm release.
func (e *DefaultExecutor) RunHelmStatus(ctx context.Context, release, namespace string) (*HelmReleaseStatus, error) {
	args := []string{"status", release, "--output", "json"}

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	result, err := e.runCommand(ctx, BaseOptions{}, "helm", args...)
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("helm status failed: %s", result.Stderr)
	}

	// Parse JSON output
	var status helmStatusOutput
	if err := json.Unmarshal([]byte(result.Stdout), &status); err != nil {
		return nil, fmt.Errorf("failed to parse helm status output: %w", err)
	}

	updated, _ := time.Parse(time.RFC3339, status.Info.LastDeployed)

	return &HelmReleaseStatus{
		Name:       status.Name,
		Namespace:  status.Namespace,
		Revision:   status.Version,
		Status:     status.Info.Status,
		Chart:      fmt.Sprintf("%s-%s", status.Chart.Metadata.Name, status.Chart.Metadata.Version),
		AppVersion: status.Chart.Metadata.AppVersion,
		Updated:    updated,
		Notes:      status.Info.Notes,
	}, nil
}

// helmStatusOutput represents the JSON output from helm status.
type helmStatusOutput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   int    `json:"version"`
	Info      struct {
		Status       string `json:"status"`
		LastDeployed string `json:"last_deployed"`
		Notes        string `json:"notes"`
	} `json:"info"`
	Chart struct {
		Metadata struct {
			Name       string `json:"name"`
			Version    string `json:"version"`
			AppVersion string `json:"appVersion"`
		} `json:"metadata"`
	} `json:"chart"`
}
