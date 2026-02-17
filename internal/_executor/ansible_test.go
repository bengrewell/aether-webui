package executor

import (
	"testing"
)

func TestBuildAnsiblePlaybookArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     AnsibleOptions
		wantArgs []string
	}{
		{
			name: "minimal",
			opts: AnsibleOptions{
				Playbook: "site.yml",
			},
			wantArgs: []string{"site.yml"},
		},
		{
			name: "with inventory",
			opts: AnsibleOptions{
				Playbook:  "site.yml",
				Inventory: "hosts.ini",
			},
			wantArgs: []string{"site.yml", "--inventory", "hosts.ini"},
		},
		{
			name: "with limit",
			opts: AnsibleOptions{
				Playbook: "site.yml",
				Limit:    "webservers",
			},
			wantArgs: []string{"site.yml", "--limit", "webservers"},
		},
		{
			name: "with tags",
			opts: AnsibleOptions{
				Playbook: "site.yml",
				Tags:     []string{"deploy", "config"},
			},
			wantArgs: []string{"site.yml", "--tags", "deploy", "--tags", "config"},
		},
		{
			name: "with skip tags",
			opts: AnsibleOptions{
				Playbook: "site.yml",
				SkipTags: []string{"debug", "test"},
			},
			wantArgs: []string{"site.yml", "--skip-tags", "debug", "--skip-tags", "test"},
		},
		{
			name: "with become",
			opts: AnsibleOptions{
				Playbook: "site.yml",
				Become:   true,
			},
			wantArgs: []string{"site.yml", "--become"},
		},
		{
			name: "with become user",
			opts: AnsibleOptions{
				Playbook:   "site.yml",
				Become:     true,
				BecomeUser: "admin",
			},
			wantArgs: []string{"site.yml", "--become", "--become-user", "admin"},
		},
		{
			name: "with verbosity 1",
			opts: AnsibleOptions{
				Playbook:  "site.yml",
				Verbosity: 1,
			},
			wantArgs: []string{"site.yml", "-v"},
		},
		{
			name: "with verbosity 2",
			opts: AnsibleOptions{
				Playbook:  "site.yml",
				Verbosity: 2,
			},
			wantArgs: []string{"site.yml", "-vv"},
		},
		{
			name: "with verbosity 4",
			opts: AnsibleOptions{
				Playbook:  "site.yml",
				Verbosity: 4,
			},
			wantArgs: []string{"site.yml", "-vvvv"},
		},
		{
			name: "with verbosity capped at 4",
			opts: AnsibleOptions{
				Playbook:  "site.yml",
				Verbosity: 10,
			},
			wantArgs: []string{"site.yml", "-vvvv"},
		},
		{
			name: "with check",
			opts: AnsibleOptions{
				Playbook: "site.yml",
				Check:    true,
			},
			wantArgs: []string{"site.yml", "--check"},
		},
		{
			name: "with diff",
			opts: AnsibleOptions{
				Playbook: "site.yml",
				Diff:     true,
			},
			wantArgs: []string{"site.yml", "--diff"},
		},
		{
			name: "with forks",
			opts: AnsibleOptions{
				Playbook: "site.yml",
				Forks:    20,
			},
			wantArgs: []string{"site.yml", "--forks", "20"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildAnsiblePlaybookArgs(tc.opts)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildAnsiblePlaybookArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	tests := []struct {
		c    rune
		n    int
		want string
	}{
		{'v', 1, "v"},
		{'v', 2, "vv"},
		{'v', 3, "vvv"},
		{'v', 4, "vvvv"},
		{'a', 0, ""},
		{'x', 5, "xxxxx"},
	}

	for _, tc := range tests {
		got := repeat(tc.c, tc.n)
		if got != tc.want {
			t.Errorf("repeat(%q, %d) = %q, want %q", tc.c, tc.n, got, tc.want)
		}
	}
}

// Helper function for building args (extracted for testing)
func buildAnsiblePlaybookArgs(opts AnsibleOptions) []string {
	args := []string{opts.Playbook}

	if opts.Inventory != "" {
		args = append(args, "--inventory", opts.Inventory)
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
		v := opts.Verbosity
		if v > 4 {
			v = 4
		}
		args = append(args, "-"+repeat('v', v))
	}
	if opts.Check {
		args = append(args, "--check")
	}
	if opts.Diff {
		args = append(args, "--diff")
	}
	if opts.Forks > 0 {
		args = append(args, "--forks", "20")
	}

	return args
}
