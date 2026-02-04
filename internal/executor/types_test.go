package executor

import (
	"testing"
	"time"
)

func TestExecResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		want     bool
	}{
		{"exit 0", 0, true},
		{"exit 1", 1, false},
		{"exit 127", 127, false},
		{"exit -1", -1, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &ExecResult{ExitCode: tc.exitCode}
			if got := r.Success(); got != tc.want {
				t.Errorf("Success() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExecResult_Fields(t *testing.T) {
	r := &ExecResult{
		ExitCode: 0,
		Stdout:   "output",
		Stderr:   "error",
		Duration: 100 * time.Millisecond,
	}

	if r.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", r.ExitCode)
	}
	if r.Stdout != "output" {
		t.Errorf("Stdout = %q, want %q", r.Stdout, "output")
	}
	if r.Stderr != "error" {
		t.Errorf("Stderr = %q, want %q", r.Stderr, "error")
	}
	if r.Duration != 100*time.Millisecond {
		t.Errorf("Duration = %v, want %v", r.Duration, 100*time.Millisecond)
	}
}

func TestBaseOptions_Fields(t *testing.T) {
	opts := BaseOptions{
		WorkingDir: "/tmp",
		Env:        map[string]string{"FOO": "bar"},
		Timeout:    5 * time.Minute,
	}

	if opts.WorkingDir != "/tmp" {
		t.Errorf("WorkingDir = %q, want %q", opts.WorkingDir, "/tmp")
	}
	if opts.Env["FOO"] != "bar" {
		t.Errorf("Env[FOO] = %q, want %q", opts.Env["FOO"], "bar")
	}
	if opts.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want %v", opts.Timeout, 5*time.Minute)
	}
}

func TestHelmInstallOptions_Fields(t *testing.T) {
	opts := HelmInstallOptions{
		ReleaseName: "my-release",
		Chart:       "my-chart",
		Namespace:   "my-ns",
		Values:      map[string]any{"key": "value"},
		ValuesFiles: []string{"values.yaml"},
		Version:     "1.0.0",
		Repo:        "https://charts.example.com",
		Wait:        true,
		WaitTimeout: 5 * time.Minute,
		CreateNS:    true,
		Atomic:      true,
		DryRun:      false,
	}

	if opts.ReleaseName != "my-release" {
		t.Errorf("ReleaseName = %q, want %q", opts.ReleaseName, "my-release")
	}
	if opts.Chart != "my-chart" {
		t.Errorf("Chart = %q, want %q", opts.Chart, "my-chart")
	}
	if !opts.Wait {
		t.Error("Wait should be true")
	}
	if !opts.CreateNS {
		t.Error("CreateNS should be true")
	}
}

func TestDockerRunOptions_Fields(t *testing.T) {
	opts := DockerRunOptions{
		Image:      "nginx:latest",
		Name:       "my-nginx",
		Command:    []string{"nginx", "-g", "daemon off;"},
		Detach:     true,
		Remove:     false,
		Network:    "bridge",
		Ports:      map[string]string{"8080": "80"},
		Volumes:    map[string]string{"/data": "/var/www"},
		EnvVars:    map[string]string{"ENV": "prod"},
		Privileged: false,
	}

	if opts.Image != "nginx:latest" {
		t.Errorf("Image = %q, want %q", opts.Image, "nginx:latest")
	}
	if !opts.Detach {
		t.Error("Detach should be true")
	}
	if opts.Ports["8080"] != "80" {
		t.Errorf("Ports[8080] = %q, want %q", opts.Ports["8080"], "80")
	}
}

func TestAnsibleOptions_Fields(t *testing.T) {
	opts := AnsibleOptions{
		Playbook:   "site.yml",
		Inventory:  "hosts.ini",
		ExtraVars:  map[string]string{"env": "prod"},
		Limit:      "webservers",
		Tags:       []string{"deploy"},
		SkipTags:   []string{"debug"},
		Become:     true,
		BecomeUser: "root",
		Verbosity:  2,
		Check:      false,
		Diff:       true,
		Forks:      10,
	}

	if opts.Playbook != "site.yml" {
		t.Errorf("Playbook = %q, want %q", opts.Playbook, "site.yml")
	}
	if !opts.Become {
		t.Error("Become should be true")
	}
	if opts.Verbosity != 2 {
		t.Errorf("Verbosity = %d, want %d", opts.Verbosity, 2)
	}
}

func TestKubectlOptions_Fields(t *testing.T) {
	opts := KubectlOptions{
		Args:       []string{"get", "pods"},
		Namespace:  "default",
		Context:    "my-cluster",
		Kubeconfig: "/home/user/.kube/config",
		Output:     "json",
	}

	if len(opts.Args) != 2 {
		t.Errorf("Args length = %d, want 2", len(opts.Args))
	}
	if opts.Namespace != "default" {
		t.Errorf("Namespace = %q, want %q", opts.Namespace, "default")
	}
	if opts.Output != "json" {
		t.Errorf("Output = %q, want %q", opts.Output, "json")
	}
}

func TestShellOptions_Fields(t *testing.T) {
	opts := ShellOptions{
		Command: "echo hello",
		Shell:   "/bin/bash",
		Args:    []string{"-x"},
	}

	if opts.Command != "echo hello" {
		t.Errorf("Command = %q, want %q", opts.Command, "echo hello")
	}
	if opts.Shell != "/bin/bash" {
		t.Errorf("Shell = %q, want %q", opts.Shell, "/bin/bash")
	}
}

func TestScriptOptions_Fields(t *testing.T) {
	opts := ScriptOptions{
		Path:        "/scripts/deploy.sh",
		Args:        []string{"--env", "prod"},
		Interpreter: "/bin/bash",
	}

	if opts.Path != "/scripts/deploy.sh" {
		t.Errorf("Path = %q, want %q", opts.Path, "/scripts/deploy.sh")
	}
	if opts.Interpreter != "/bin/bash" {
		t.Errorf("Interpreter = %q, want %q", opts.Interpreter, "/bin/bash")
	}
}

func TestHelmRelease_Fields(t *testing.T) {
	now := time.Now()
	r := HelmRelease{
		Name:       "my-release",
		Namespace:  "default",
		Revision:   3,
		Status:     "deployed",
		Chart:      "nginx-1.0.0",
		AppVersion: "1.19.0",
		Updated:    now,
	}

	if r.Name != "my-release" {
		t.Errorf("Name = %q, want %q", r.Name, "my-release")
	}
	if r.Revision != 3 {
		t.Errorf("Revision = %d, want %d", r.Revision, 3)
	}
	if r.Status != "deployed" {
		t.Errorf("Status = %q, want %q", r.Status, "deployed")
	}
}

func TestHelmReleaseList_Fields(t *testing.T) {
	list := HelmReleaseList{
		Releases: []HelmRelease{
			{Name: "release-1"},
			{Name: "release-2"},
		},
	}

	if len(list.Releases) != 2 {
		t.Errorf("Releases length = %d, want 2", len(list.Releases))
	}
}

func TestHelmReleaseStatus_Fields(t *testing.T) {
	status := HelmReleaseStatus{
		Name:       "my-release",
		Namespace:  "default",
		Revision:   1,
		Status:     "deployed",
		Chart:      "nginx-1.0.0",
		AppVersion: "1.19.0",
		Notes:      "Installation notes",
		Manifest:   "---\napiVersion: v1\n",
	}

	if status.Notes != "Installation notes" {
		t.Errorf("Notes = %q, want %q", status.Notes, "Installation notes")
	}
}
