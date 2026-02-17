package executor

import (
	"context"
)

// RunKubectl runs a generic kubectl command.
func (e *DefaultExecutor) RunKubectl(ctx context.Context, opts KubectlOptions) (*ExecResult, error) {
	args := make([]string, 0, len(opts.Args)+6)

	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	if opts.Context != "" {
		args = append(args, "--context", opts.Context)
	}
	if opts.Kubeconfig != "" {
		args = append(args, "--kubeconfig", opts.Kubeconfig)
	}
	if opts.Output != "" {
		args = append(args, "--output", opts.Output)
	}

	args = append(args, opts.Args...)

	return e.runCommand(ctx, opts.BaseOptions, "kubectl", args...)
}

// KubectlApply applies a manifest to the cluster.
func (e *DefaultExecutor) KubectlApply(ctx context.Context, manifest []byte, namespace string) (*ExecResult, error) {
	args := []string{"apply", "-f", "-"}

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	return e.runCommandWithStdin(ctx, BaseOptions{}, manifest, "kubectl", args...)
}

// KubectlDelete deletes a resource from the cluster.
func (e *DefaultExecutor) KubectlDelete(ctx context.Context, resource, name, namespace string) (*ExecResult, error) {
	args := []string{"delete", resource, name}

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	return e.runCommand(ctx, BaseOptions{}, "kubectl", args...)
}

// KubectlGet gets a resource from the cluster.
func (e *DefaultExecutor) KubectlGet(ctx context.Context, resource, name, namespace string, output string) (*ExecResult, error) {
	args := []string{"get", resource}

	if name != "" {
		args = append(args, name)
	}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	if output != "" {
		args = append(args, "--output", output)
	}

	return e.runCommand(ctx, BaseOptions{}, "kubectl", args...)
}
