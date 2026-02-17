package executor

import (
	"context"
	"fmt"
	"time"
)

// RunDockerCommand runs a generic Docker command.
func (e *DefaultExecutor) RunDockerCommand(ctx context.Context, opts DockerOptions) (*ExecResult, error) {
	base := opts.BaseOptions
	if opts.Host != "" {
		if base.Env == nil {
			base.Env = make(map[string]string)
		}
		base.Env["DOCKER_HOST"] = opts.Host
	}

	return e.runCommand(ctx, base, "docker", opts.Args...)
}

// DockerRun runs a Docker container.
func (e *DefaultExecutor) DockerRun(ctx context.Context, opts DockerRunOptions) (*ExecResult, error) {
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
	for hostPort, containerPort := range opts.Ports {
		args = append(args, "--publish", fmt.Sprintf("%s:%s", hostPort, containerPort))
	}
	for hostPath, containerPath := range opts.Volumes {
		args = append(args, "--volume", fmt.Sprintf("%s:%s", hostPath, containerPath))
	}
	for k, v := range opts.EnvVars {
		args = append(args, "--env", fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range opts.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
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

	base := opts.BaseOptions
	if opts.Host != "" {
		if base.Env == nil {
			base.Env = make(map[string]string)
		}
		base.Env["DOCKER_HOST"] = opts.Host
	}

	return e.runCommand(ctx, base, "docker", args...)
}

// DockerStop stops a running container.
func (e *DefaultExecutor) DockerStop(ctx context.Context, container string, timeout time.Duration) (*ExecResult, error) {
	args := []string{"stop"}

	if timeout > 0 {
		args = append(args, "--time", fmt.Sprintf("%d", int(timeout.Seconds())))
	}

	args = append(args, container)

	return e.runCommand(ctx, BaseOptions{}, "docker", args...)
}

// DockerRemove removes a container.
func (e *DefaultExecutor) DockerRemove(ctx context.Context, container string, force bool) (*ExecResult, error) {
	args := []string{"rm"}

	if force {
		args = append(args, "--force")
	}

	args = append(args, container)

	return e.runCommand(ctx, BaseOptions{}, "docker", args...)
}
