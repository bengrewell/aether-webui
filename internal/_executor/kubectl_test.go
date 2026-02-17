package executor

import (
	"testing"
)

func TestBuildKubectlArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     KubectlOptions
		wantArgs []string
	}{
		{
			name: "minimal",
			opts: KubectlOptions{
				Args: []string{"get", "pods"},
			},
			wantArgs: []string{"get", "pods"},
		},
		{
			name: "with namespace",
			opts: KubectlOptions{
				Args:      []string{"get", "pods"},
				Namespace: "my-ns",
			},
			wantArgs: []string{"--namespace", "my-ns", "get", "pods"},
		},
		{
			name: "with context",
			opts: KubectlOptions{
				Args:    []string{"get", "pods"},
				Context: "my-cluster",
			},
			wantArgs: []string{"--context", "my-cluster", "get", "pods"},
		},
		{
			name: "with kubeconfig",
			opts: KubectlOptions{
				Args:       []string{"get", "pods"},
				Kubeconfig: "/home/user/.kube/config",
			},
			wantArgs: []string{"--kubeconfig", "/home/user/.kube/config", "get", "pods"},
		},
		{
			name: "with output",
			opts: KubectlOptions{
				Args:   []string{"get", "pods"},
				Output: "json",
			},
			wantArgs: []string{"--output", "json", "get", "pods"},
		},
		{
			name: "all options",
			opts: KubectlOptions{
				Args:       []string{"get", "pods"},
				Namespace:  "my-ns",
				Context:    "my-cluster",
				Kubeconfig: "/home/user/.kube/config",
				Output:     "yaml",
			},
			wantArgs: []string{
				"--namespace", "my-ns",
				"--context", "my-cluster",
				"--kubeconfig", "/home/user/.kube/config",
				"--output", "yaml",
				"get", "pods",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildKubectlArgs(tc.opts)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildKubectlArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildKubectlApplyArgs(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantArgs  []string
	}{
		{
			name:      "no namespace",
			namespace: "",
			wantArgs:  []string{"apply", "-f", "-"},
		},
		{
			name:      "with namespace",
			namespace: "my-ns",
			wantArgs:  []string{"apply", "-f", "-", "--namespace", "my-ns"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildKubectlApplyArgs(tc.namespace)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildKubectlApplyArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildKubectlDeleteArgs(t *testing.T) {
	tests := []struct {
		name      string
		resource  string
		resName   string
		namespace string
		wantArgs  []string
	}{
		{
			name:      "no namespace",
			resource:  "pod",
			resName:   "my-pod",
			namespace: "",
			wantArgs:  []string{"delete", "pod", "my-pod"},
		},
		{
			name:      "with namespace",
			resource:  "deployment",
			resName:   "my-deploy",
			namespace: "my-ns",
			wantArgs:  []string{"delete", "deployment", "my-deploy", "--namespace", "my-ns"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildKubectlDeleteArgs(tc.resource, tc.resName, tc.namespace)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildKubectlDeleteArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

func TestBuildKubectlGetArgs(t *testing.T) {
	tests := []struct {
		name      string
		resource  string
		resName   string
		namespace string
		output    string
		wantArgs  []string
	}{
		{
			name:      "list all",
			resource:  "pods",
			resName:   "",
			namespace: "",
			output:    "",
			wantArgs:  []string{"get", "pods"},
		},
		{
			name:      "specific resource",
			resource:  "pod",
			resName:   "my-pod",
			namespace: "",
			output:    "",
			wantArgs:  []string{"get", "pod", "my-pod"},
		},
		{
			name:      "with namespace",
			resource:  "pods",
			resName:   "",
			namespace: "my-ns",
			output:    "",
			wantArgs:  []string{"get", "pods", "--namespace", "my-ns"},
		},
		{
			name:      "with output",
			resource:  "pods",
			resName:   "",
			namespace: "",
			output:    "json",
			wantArgs:  []string{"get", "pods", "--output", "json"},
		},
		{
			name:      "all options",
			resource:  "pod",
			resName:   "my-pod",
			namespace: "my-ns",
			output:    "yaml",
			wantArgs:  []string{"get", "pod", "my-pod", "--namespace", "my-ns", "--output", "yaml"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := buildKubectlGetArgs(tc.resource, tc.resName, tc.namespace, tc.output)
			if !slicesEqual(args, tc.wantArgs) {
				t.Errorf("buildKubectlGetArgs() = %v, want %v", args, tc.wantArgs)
			}
		})
	}
}

// Helper functions for building args (extracted for testing)
func buildKubectlArgs(opts KubectlOptions) []string {
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
	return args
}

func buildKubectlApplyArgs(namespace string) []string {
	args := []string{"apply", "-f", "-"}

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	return args
}

func buildKubectlDeleteArgs(resource, name, namespace string) []string {
	args := []string{"delete", resource, name}

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	return args
}

func buildKubectlGetArgs(resource, name, namespace string, output string) []string {
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

	return args
}
