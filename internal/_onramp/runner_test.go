package onramp

import (
	"path/filepath"
	"testing"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner("/opt/onramp")
	if r.onrampPath != "/opt/onramp" {
		t.Errorf("onrampPath = %q, want %q", r.onrampPath, "/opt/onramp")
	}
	if r.varsFile != "/opt/onramp/vars/main.yml" {
		t.Errorf("varsFile = %q, want %q", r.varsFile, "/opt/onramp/vars/main.yml")
	}
}

func TestOnRampPath(t *testing.T) {
	r := NewRunner("/opt/onramp")
	if got := r.OnRampPath(); got != "/opt/onramp" {
		t.Errorf("OnRampPath() = %q, want %q", got, "/opt/onramp")
	}
}

func TestBuildEnvVars(t *testing.T) {
	r := NewRunner("/opt/onramp")
	env := r.buildEnvVars("/opt/onramp/hosts.ini")

	expected := map[string]string{
		"ANSIBLE_CONFIG":    "/opt/onramp/ansible.cfg",
		"ROOT_DIR":          "/opt/onramp",
		"AETHER_ROOT_DIR":   "/opt/onramp",
		"HOSTS_INI_FILE":    "/opt/onramp/hosts.ini",
		"5GC_ROOT_DIR":      filepath.Join("/opt/onramp", "deps", "5gc"),
		"K8S_ROOT_DIR":      filepath.Join("/opt/onramp", "deps", "k8s"),
		"SRSRAN_ROOT_DIR":   filepath.Join("/opt/onramp", "deps", "srsran"),
		"GNBSIM_ROOT_DIR":   filepath.Join("/opt/onramp", "deps", "gnbsim"),
		"AMP_ROOT_DIR":      filepath.Join("/opt/onramp", "deps", "amp"),
		"OAI_ROOT_DIR":      filepath.Join("/opt/onramp", "deps", "oai"),
		"SDRAN_ROOT_DIR":    filepath.Join("/opt/onramp", "deps", "sdran"),
		"UERANSIM_ROOT_DIR": filepath.Join("/opt/onramp", "deps", "ueransim"),
		"OSCRIC_ROOT_DIR":   filepath.Join("/opt/onramp", "deps", "oscric"),
		"N3IWF_ROOT_DIR":    filepath.Join("/opt/onramp", "deps", "n3iwf"),
		"4GC_ROOT_DIR":      filepath.Join("/opt/onramp", "deps", "4gc"),
	}

	for key, want := range expected {
		got, ok := env[key]
		if !ok {
			t.Errorf("missing env var %s", key)
			continue
		}
		if got != want {
			t.Errorf("env[%s] = %q, want %q", key, got, want)
		}
	}

	if len(env) != len(expected) {
		t.Errorf("env has %d vars, want %d", len(env), len(expected))
	}
}
