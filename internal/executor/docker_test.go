package executor

import (
	"testing"
	"time"
)

func TestBuildDockerRunArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     DockerRunOptions
		wantArgs []string
	}{
		{
			name: "minimal",
			opts: DockerRunOptions{
				Image: "nginx:latest",
			},
			wantArgs: []string{"run", "nginx:latest"},
		},
		{
			name: "with name",
			opts: DockerRunOptions{
				Image: "nginx:latest",
				Name:  "my-nginx",
			},
			wantArgs: []string{"run", "--name", "my-nginx", "nginx:latest"},
		},
		{
			name: "with detach",
			opts: DockerRunOptions{
				Image:  "nginx:latest",
				Detach: true,
			},
			wantArgs: []string{"run", "--detach", "nginx:latest"},
		},
		{
			name: "with rm",
			opts: DockerRunOptions{
				Image:  "nginx:latest",
				Remove: true,
			},
			wantArgs: []string{"run", "--rm", "nginx:latest"},
		},
		{
			name: "with interactive tty",
			opts: DockerRunOptions{
				Image:       "ubuntu:latest",
				Interactive: true,
				TTY:         true,
			},
			wantArgs: []string{"run", "--interactive", "--tty", "ubuntu:latest"},
		},
		{
			name: "with privileged",
			opts: DockerRunOptions{
				Image:      "nginx:latest",
				Privileged: true,
			},
			wantArgs: []string{"run", "--privileged", "nginx:latest"},
		},
		{
			name: "with network",
			opts: DockerRunOptions{
				Image:   "nginx:latest",
				Network: "host",
			},
			wantArgs: []string{"run", "--network", "host", "nginx:latest"},
		},
		{
			name: "with user",
			opts: DockerRunOptions{
				Image: "nginx:latest",
				User:  "1000:1000",
			},
			wantArgs: []string{"run", "--user", "1000:1000", "nginx:latest"},
		},
		{
			name: "with hostname",
			opts: DockerRunOptions{
				Image:    "nginx:latest",
				Hostname: "myhost",
			},
			wantArgs: []string{"run", "--hostname", "myhost", "nginx:latest"},
		},
		{
			name: "with restart policy",
			opts: DockerRunOptions{
				Image:         "nginx:latest",
				RestartPolicy: "always",
			},
			wantArgs: []string{"run", "--restart", "always", "nginx:latest"},
		},
		{
			name: "with memory limit",
			opts: DockerRunOptions{
				Image:  "nginx:latest",
				Memory: "512m",
			},
			wantArgs: []string{"run", "--memory", "512m", "nginx:latest"},
		},
		{
			name: "with cpu limit",
			opts: DockerRunOptions{
				Image: "nginx:latest",
				CPUs:  "0.5",
			},
			wantArgs: []string{"run", "--cpus", "0.5", "nginx:latest"},
		},
		{
			name: "with command",
			opts: DockerRunOptions{
				Image:   "nginx:latest",
				Command: []string{"nginx", "-g", "daemon off;"},
			},
			wantArgs: []string{"run", "nginx:latest", "nginx", "-g", "daemon off;"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildDockerRunArgs(tc.opts)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildDockerRunArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildDockerStopArgs(t *testing.T) {
	tests := []struct {
		name      string
		container string
		timeout   time.Duration
		wantArgs  []string
	}{
		{
			name:      "no timeout",
			container: "my-container",
			timeout:   0,
			wantArgs:  []string{"stop", "my-container"},
		},
		{
			name:      "with timeout",
			container: "my-container",
			timeout:   30 * time.Second,
			wantArgs:  []string{"stop", "--time", "30", "my-container"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildDockerStopArgs(tc.container, tc.timeout)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildDockerStopArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildDockerRemoveArgs(t *testing.T) {
	tests := []struct {
		name      string
		container string
		force     bool
		wantArgs  []string
	}{
		{
			name:      "without force",
			container: "my-container",
			force:     false,
			wantArgs:  []string{"rm", "my-container"},
		},
		{
			name:      "with force",
			container: "my-container",
			force:     true,
			wantArgs:  []string{"rm", "--force", "my-container"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildDockerRemoveArgs(tc.container, tc.force)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildDockerRemoveArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

// Helper functions for building args (extracted for testing)
func buildDockerRunArgs(opts DockerRunOptions) []string {
	args := []string{"run"}

	if opts.Name != "" {
		args = append(args, "--name", opts.Name)
	}
	if opts.Detach {
		args = append(args, "--detach")
	}
	if opts.Remove {
		args = append(args, "--rm")
	}
	if opts.Interactive {
		args = append(args, "--interactive")
	}
	if opts.TTY {
		args = append(args, "--tty")
	}
	if opts.Privileged {
		args = append(args, "--privileged")
	}
	if opts.Network != "" {
		args = append(args, "--network", opts.Network)
	}
	if opts.User != "" {
		args = append(args, "--user", opts.User)
	}
	if opts.Hostname != "" {
		args = append(args, "--hostname", opts.Hostname)
	}
	if opts.RestartPolicy != "" {
		args = append(args, "--restart", opts.RestartPolicy)
	}
	if opts.Memory != "" {
		args = append(args, "--memory", opts.Memory)
	}
	if opts.CPUs != "" {
		args = append(args, "--cpus", opts.CPUs)
	}

	args = append(args, opts.Image)
	args = append(args, opts.Command...)

	return args
}

func buildDockerStopArgs(container string, timeout time.Duration) []string {
	args := []string{"stop"}

	if timeout > 0 {
		args = append(args, "--time", "30")
	}

	args = append(args, container)
	return args
}

func buildDockerRemoveArgs(container string, force bool) []string {
	args := []string{"rm"}

	if force {
		args = append(args, "--force")
	}

	args = append(args, container)
	return args
}
