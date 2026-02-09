package onramp

import (
	"testing"
)

func TestAllSequencesDefined(t *testing.T) {
	expected := []string{
		"aether-pingall",
		"k8s-install",
		"k8s-uninstall",
		"5gc-install",
		"5gc-uninstall",
		"srsran-gnb-install",
		"srsran-gnb-uninstall",
	}

	for _, name := range expected {
		if _, ok := Sequences[name]; !ok {
			t.Errorf("missing sequence %q", name)
		}
	}
}

func TestSequencesHaveSteps(t *testing.T) {
	for name, seq := range Sequences {
		if len(seq.Steps) == 0 {
			t.Errorf("sequence %q has no steps", name)
		}
		if seq.Name == "" {
			t.Errorf("sequence %q has no display name", name)
		}
		for i, step := range seq.Steps {
			if step.Playbook == "" {
				t.Errorf("sequence %q step %d has empty playbook", name, i)
			}
			if step.Name == "" {
				t.Errorf("sequence %q step %d has empty name", name, i)
			}
		}
	}
}

func TestInstallOrderRouterBeforeCore(t *testing.T) {
	seq := Sequences["5gc-install"]
	if len(seq.Steps) < 2 {
		t.Fatal("5gc-install should have at least 2 steps")
	}
	if seq.Steps[0].Playbook != "deps/5gc/router.yml" {
		t.Errorf("first step should be router, got %s", seq.Steps[0].Playbook)
	}
	if seq.Steps[1].Playbook != "deps/5gc/core.yml" {
		t.Errorf("second step should be core, got %s", seq.Steps[1].Playbook)
	}
}

func TestUninstallOrderCoreBeforeRouter(t *testing.T) {
	seq := Sequences["5gc-uninstall"]
	if len(seq.Steps) < 2 {
		t.Fatal("5gc-uninstall should have at least 2 steps")
	}
	if seq.Steps[0].Playbook != "deps/5gc/core.yml" {
		t.Errorf("first step should be core, got %s", seq.Steps[0].Playbook)
	}
	if seq.Steps[1].Playbook != "deps/5gc/router.yml" {
		t.Errorf("second step should be router, got %s", seq.Steps[1].Playbook)
	}
}

func TestTagString(t *testing.T) {
	step := PlaybookStep{Tags: []string{"install"}}
	if got := step.TagString(); got != "install" {
		t.Errorf("TagString() = %q, want %q", got, "install")
	}

	step2 := PlaybookStep{}
	if got := step2.TagString(); got != "" {
		t.Errorf("TagString() = %q, want empty", got)
	}
}
