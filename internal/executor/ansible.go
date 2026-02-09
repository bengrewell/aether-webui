package executor

import (
	"context"
	"fmt"
	"strconv"
)

// RunAnsiblePlaybook executes an Ansible playbook.
func (e *DefaultExecutor) RunAnsiblePlaybook(ctx context.Context, opts AnsibleOptions) (*ExecResult, error) {
	args := []string{opts.Playbook}

	if opts.Inventory != "" {
		args = append(args, "--inventory", opts.Inventory)
	}
	for k, v := range opts.ExtraVars {
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", k, v))
	}
	if opts.Limit != "" {
		args = append(args, "--limit", opts.Limit)
	}
	for _, tag := range opts.Tags {
		args = append(args, "--tags", tag)
	}
	for _, tag := range opts.SkipTags {
		args = append(args, "--skip-tags", tag)
	}
	if opts.Become {
		args = append(args, "--become")
		if opts.BecomeUser != "" {
			args = append(args, "--become-user", opts.BecomeUser)
		}
	}
	if opts.Verbosity > 0 {
		// Convert verbosity level to -v, -vv, -vvv, -vvvv
		v := opts.Verbosity
		if v > 4 {
			v = 4
		}
		args = append(args, "-"+string(repeat('v', v)))
	}
	if opts.Check {
		args = append(args, "--check")
	}
	if opts.Diff {
		args = append(args, "--diff")
	}
	if opts.Forks > 0 {
		args = append(args, "--forks", strconv.Itoa(opts.Forks))
	}

	return e.runCommand(ctx, opts.BaseOptions, "ansible-playbook", args...)
}

// repeat returns a string with the character c repeated n times.
func repeat(c rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}
