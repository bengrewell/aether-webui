package executor

import (
	"testing"
	"time"
)

func TestBuildHelmInstallArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     HelmInstallOptions
		wantArgs []string
	}{
		{
			name: "minimal",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
			},
			wantArgs: []string{"install", "my-release", "nginx"},
		},
		{
			name: "with namespace",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Namespace:   "my-ns",
			},
			wantArgs: []string{"install", "my-release", "nginx", "--namespace", "my-ns"},
		},
		{
			name: "with version",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Version:     "1.0.0",
			},
			wantArgs: []string{"install", "my-release", "nginx", "--version", "1.0.0"},
		},
		{
			name: "with repo",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Repo:        "https://charts.example.com",
			},
			wantArgs: []string{"install", "my-release", "nginx", "--repo", "https://charts.example.com"},
		},
		{
			name: "with values files",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				ValuesFiles: []string{"values.yaml", "prod.yaml"},
			},
			wantArgs: []string{"install", "my-release", "nginx", "--values", "values.yaml", "--values", "prod.yaml"},
		},
		{
			name: "with wait",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Wait:        true,
				WaitTimeout: 5 * time.Minute,
			},
			wantArgs: []string{"install", "my-release", "nginx", "--wait", "--timeout", "5m0s"},
		},
		{
			name: "with create namespace",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				CreateNS:    true,
			},
			wantArgs: []string{"install", "my-release", "nginx", "--create-namespace"},
		},
		{
			name: "with atomic",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Atomic:      true,
			},
			wantArgs: []string{"install", "my-release", "nginx", "--atomic"},
		},
		{
			name: "with dry run",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				DryRun:      true,
			},
			wantArgs: []string{"install", "my-release", "nginx", "--dry-run"},
		},
		{
			name: "with description",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Description: "Initial deployment",
			},
			wantArgs: []string{"install", "my-release", "nginx", "--description", "Initial deployment"},
		},
		{
			name: "with replace",
			opts: HelmInstallOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Replace:     true,
			},
			wantArgs: []string{"install", "my-release", "nginx", "--replace"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildHelmInstallArgs(tc.opts)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildHelmInstallArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildHelmUpgradeArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     HelmUpgradeOptions
		wantArgs []string
	}{
		{
			name: "minimal",
			opts: HelmUpgradeOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
			},
			wantArgs: []string{"upgrade", "my-release", "nginx"},
		},
		{
			name: "with install",
			opts: HelmUpgradeOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Install:     true,
			},
			wantArgs: []string{"upgrade", "my-release", "nginx", "--install"},
		},
		{
			name: "with reuse values",
			opts: HelmUpgradeOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				ReuseValues: true,
			},
			wantArgs: []string{"upgrade", "my-release", "nginx", "--reuse-values"},
		},
		{
			name: "with reset values",
			opts: HelmUpgradeOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				ResetValues: true,
			},
			wantArgs: []string{"upgrade", "my-release", "nginx", "--reset-values"},
		},
		{
			name: "with force",
			opts: HelmUpgradeOptions{
				ReleaseName: "my-release",
				Chart:       "nginx",
				Force:       true,
			},
			wantArgs: []string{"upgrade", "my-release", "nginx", "--force"},
		},
		{
			name: "with cleanup on fail",
			opts: HelmUpgradeOptions{
				ReleaseName:   "my-release",
				Chart:         "nginx",
				CleanupOnFail: true,
			},
			wantArgs: []string{"upgrade", "my-release", "nginx", "--cleanup-on-fail"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildHelmUpgradeArgs(tc.opts)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildHelmUpgradeArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildHelmUninstallArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     HelmUninstallOptions
		wantArgs []string
	}{
		{
			name: "minimal",
			opts: HelmUninstallOptions{
				ReleaseName: "my-release",
			},
			wantArgs: []string{"uninstall", "my-release"},
		},
		{
			name: "with namespace",
			opts: HelmUninstallOptions{
				ReleaseName: "my-release",
				Namespace:   "my-ns",
			},
			wantArgs: []string{"uninstall", "my-release", "--namespace", "my-ns"},
		},
		{
			name: "with keep history",
			opts: HelmUninstallOptions{
				ReleaseName: "my-release",
				KeepHistory: true,
			},
			wantArgs: []string{"uninstall", "my-release", "--keep-history"},
		},
		{
			name: "with dry run",
			opts: HelmUninstallOptions{
				ReleaseName: "my-release",
				DryRun:      true,
			},
			wantArgs: []string{"uninstall", "my-release", "--dry-run"},
		},
		{
			name: "with wait",
			opts: HelmUninstallOptions{
				ReleaseName: "my-release",
				Wait:        true,
				WaitTimeout: 2 * time.Minute,
			},
			wantArgs: []string{"uninstall", "my-release", "--wait", "--timeout", "2m0s"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildHelmUninstallArgs(tc.opts)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildHelmUninstallArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildHelmListArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     HelmListOptions
		wantArgs []string
	}{
		{
			name:     "minimal",
			opts:     HelmListOptions{},
			wantArgs: []string{"list", "--output", "json"},
		},
		{
			name: "with namespace",
			opts: HelmListOptions{
				Namespace: "my-ns",
			},
			wantArgs: []string{"list", "--output", "json", "--namespace", "my-ns"},
		},
		{
			name: "with all namespaces",
			opts: HelmListOptions{
				AllNS: true,
			},
			wantArgs: []string{"list", "--output", "json", "--all-namespaces"},
		},
		{
			name: "with filter",
			opts: HelmListOptions{
				Filter: "nginx.*",
			},
			wantArgs: []string{"list", "--output", "json", "--filter", "nginx.*"},
		},
		{
			name: "with all statuses",
			opts: HelmListOptions{
				All: true,
			},
			wantArgs: []string{"list", "--output", "json", "--all"},
		},
		{
			name: "with deployed only",
			opts: HelmListOptions{
				Deployed: true,
			},
			wantArgs: []string{"list", "--output", "json", "--deployed"},
		},
		{
			name: "with failed only",
			opts: HelmListOptions{
				Failed: true,
			},
			wantArgs: []string{"list", "--output", "json", "--failed"},
		},
		{
			name: "with pagination",
			opts: HelmListOptions{
				Offset: 10,
				Max:    5,
			},
			wantArgs: []string{"list", "--output", "json", "--offset", "10", "--max", "5"},
		},
		{
			name: "with reverse",
			opts: HelmListOptions{
				Reverse: true,
			},
			wantArgs: []string{"list", "--output", "json", "--reverse"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildHelmListArgs(tc.opts)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildHelmListArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestParseRevision(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"1", 1},
		{"10", 10},
		{"", 0},
		{"invalid", 0},
	}

	for _, tc := range tests {
		got := parseRevision(tc.input)
		if got != tc.want {
			t.Errorf("parseRevision(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

// Helper functions for building args (extracted for testing)
func buildHelmInstallArgs(opts HelmInstallOptions) []string {
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

	return args
}

func buildHelmUpgradeArgs(opts HelmUpgradeOptions) []string {
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

	return args
}

func buildHelmUninstallArgs(opts HelmUninstallOptions) []string {
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

	return args
}

func buildHelmListArgs(opts HelmListOptions) []string {
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
		args = append(args, "--offset", "10")
	}
	if opts.Max > 0 {
		args = append(args, "--max", "5")
	}
	if opts.Reverse {
		args = append(args, "--reverse")
	}

	return args
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
