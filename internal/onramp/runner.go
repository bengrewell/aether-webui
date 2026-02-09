package onramp

import (
	"context"
	"io"
	"path/filepath"

	"github.com/apenella/go-ansible/v2/pkg/execute"
	"github.com/apenella/go-ansible/v2/pkg/playbook"
)

// Runner executes ansible-playbook commands against an OnRamp checkout.
type Runner struct {
	onrampPath string
	varsFile   string
}

// NewRunner creates a Runner for the given OnRamp checkout path.
func NewRunner(onrampPath string) *Runner {
	return &Runner{
		onrampPath: onrampPath,
		varsFile:   filepath.Join(onrampPath, "vars", "main.yml"),
	}
}

// RunPlaybook executes a single playbook step with the given inventory file,
// streaming output to the provided writer.
func (r *Runner) RunPlaybook(ctx context.Context, step PlaybookStep, inventory string, output io.Writer) error {
	opts := &playbook.AnsiblePlaybookOptions{
		Inventory: inventory,
		ExtraVars: map[string]interface{}{
			"ROOT_DIR": r.onrampPath,
		},
		ExtraVarsFile: []string{"@" + r.varsFile},
	}
	if tags := step.TagString(); tags != "" {
		opts.Tags = tags
	}

	cmd := playbook.NewAnsiblePlaybookCmd(
		playbook.WithPlaybooks(step.Playbook),
		playbook.WithPlaybookOptions(opts),
	)

	exec := execute.NewDefaultExecute(
		execute.WithCmd(cmd),
		execute.WithCmdRunDir(r.onrampPath),
		execute.WithWrite(output),
		execute.WithEnvVars(r.buildEnvVars(inventory)),
	)

	return exec.Execute(ctx)
}

// buildEnvVars creates the environment variables that replicate what the
// OnRamp Makefile exports when running playbooks.
func (r *Runner) buildEnvVars(inventory string) map[string]string {
	return map[string]string{
		"ANSIBLE_CONFIG":    filepath.Join(r.onrampPath, "ansible.cfg"),
		"ROOT_DIR":          r.onrampPath,
		"AETHER_ROOT_DIR":   r.onrampPath,
		"HOSTS_INI_FILE":    inventory,
		"5GC_ROOT_DIR":      filepath.Join(r.onrampPath, "deps", "5gc"),
		"4GC_ROOT_DIR":      filepath.Join(r.onrampPath, "deps", "4gc"),
		"K8S_ROOT_DIR":      filepath.Join(r.onrampPath, "deps", "k8s"),
		"AMP_ROOT_DIR":      filepath.Join(r.onrampPath, "deps", "amp"),
		"GNBSIM_ROOT_DIR":   filepath.Join(r.onrampPath, "deps", "gnbsim"),
		"SRSRAN_ROOT_DIR":   filepath.Join(r.onrampPath, "deps", "srsran"),
		"OAI_ROOT_DIR":      filepath.Join(r.onrampPath, "deps", "oai"),
		"SDRAN_ROOT_DIR":    filepath.Join(r.onrampPath, "deps", "sdran"),
		"UERANSIM_ROOT_DIR": filepath.Join(r.onrampPath, "deps", "ueransim"),
		"OSCRIC_ROOT_DIR":   filepath.Join(r.onrampPath, "deps", "oscric"),
		"N3IWF_ROOT_DIR":    filepath.Join(r.onrampPath, "deps", "n3iwf"),
	}
}

// OnRampPath returns the configured OnRamp checkout path.
func (r *Runner) OnRampPath() string {
	return r.onrampPath
}
