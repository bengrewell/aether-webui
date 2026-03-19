package onramp

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"gopkg.in/yaml.v3"
)

// writeBlueprint writes a blueprint YAML file into the provider's vars directory.
func writeBlueprint(t *testing.T, p *OnRamp, filename string, content map[string]any) {
	t.Helper()
	data, err := yaml.Marshal(content)
	if err != nil {
		t.Fatalf("marshal blueprint: %v", err)
	}
	dir := filepath.Join(p.config.OnRampDir, "vars")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func baseConfig() string {
	return `
k8s:
  rke2:
    version: "v1.28.4"
core:
  data_iface: "eth0"
  ran_subnet: "192.168.0.0/24"
gnbsim:
  router:
    data_iface: "eth0"
amp:
  roc_models: "aether-2.1"
`
}

func TestComposeConfig_K8sAnd5gc(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	// k8s and core should be present.
	if out.Body.Config.K8s == nil {
		t.Error("expected k8s section to be present")
	}
	if out.Body.Config.Core == nil {
		t.Error("expected core section to be present")
	}
	// gnbsim and amp should be pruned.
	if out.Body.Config.GNBSim != nil {
		t.Error("expected gnbsim section to be pruned")
	}
	if out.Body.Config.AMP != nil {
		t.Error("expected amp section to be pruned")
	}
}

func TestComposeConfig_WithSRSRanBlueprint(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	writeBlueprint(t, p, "main-srsran.yml", map[string]any{
		"srsran": map[string]any{
			"docker": map[string]any{
				"container": map[string]any{
					"gnb_image": "srsran/gnb:latest",
				},
			},
		},
		"core": map[string]any{
			"ran_subnet": "",
		},
	})

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "srsran"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	if out.Body.Config.SRSRan == nil {
		t.Fatal("expected srsran section to be present")
	}
	if out.Body.Config.SRSRan.Docker == nil || out.Body.Config.SRSRan.Docker.Container == nil {
		t.Fatal("expected srsran.docker.container to be present")
	}
	if out.Body.Config.SRSRan.Docker.Container.GNBImage != "srsran/gnb:latest" {
		t.Errorf("srsran gnb_image = %q, want %q", out.Body.Config.SRSRan.Docker.Container.GNBImage, "srsran/gnb:latest")
	}
	// core.ran_subnet should be overwritten by blueprint.
	if out.Body.Config.Core == nil {
		t.Fatal("expected core section")
	}
	if out.Body.Config.Core.RANSubnet != "" {
		t.Errorf("core.ran_subnet = %q, want empty", out.Body.Config.Core.RANSubnet)
	}
	// gnbsim should be pruned.
	if out.Body.Config.GNBSim != nil {
		t.Error("expected gnbsim section to be pruned")
	}
}

func TestComposeConfig_MultipleRANComponents(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	writeBlueprint(t, p, "main-srsran.yml", map[string]any{
		"srsran": map[string]any{
			"simulation": true,
		},
	})
	writeBlueprint(t, p, "main-ueransim.yml", map[string]any{
		"ueransim": map[string]any{
			"gnb": map[string]any{
				"ip": "10.0.0.99",
			},
		},
	})

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "srsran", "ueransim"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	if out.Body.Config.SRSRan == nil {
		t.Error("expected srsran section")
	}
	if out.Body.Config.UERANSIM == nil {
		t.Error("expected ueransim section")
	}
	if out.Body.Config.UERANSIM != nil && out.Body.Config.UERANSIM.GNB != nil {
		if out.Body.Config.UERANSIM.GNB.IP != "10.0.0.99" {
			t.Errorf("ueransim.gnb.ip = %q, want %q", out.Body.Config.UERANSIM.GNB.IP, "10.0.0.99")
		}
	}
}

func TestComposeConfig_WithAMP(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "amp"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	if out.Body.Config.AMP == nil {
		t.Error("expected amp section to be retained")
	}
	if out.Body.Config.GNBSim != nil {
		t.Error("expected gnbsim to be pruned")
	}
}

func TestComposeConfig_ImplicitDeps(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	writeBlueprint(t, p, "main-srsran.yml", map[string]any{
		"srsran": map[string]any{
			"simulation": true,
		},
	})

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"5gc", "srsran"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	// k8s should be implicitly included via 5gc dependency.
	if out.Body.Config.K8s == nil {
		t.Error("expected k8s section to be present via implicit dep")
	}

	// Verify k8s is in the returned components list.
	found := false
	for _, c := range out.Body.Components {
		if c == "k8s" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected k8s in components list via implicit dep")
	}
}

func TestComposeConfig_UnknownComponent(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	_, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "nonexistent"}},
	})
	if err == nil {
		t.Fatal("expected error for unknown component")
	}
}

func TestComposeConfig_MissingBlueprint(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	// Request srsran without writing its blueprint file.
	_, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "srsran"}},
	})
	if err == nil {
		t.Fatal("expected error for missing blueprint file")
	}
}

func TestComposeConfig_WritesMainYML(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	_, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	// Re-read main.yml and verify gnbsim is gone.
	mainPath := filepath.Join(p.config.OnRampDir, "vars", "main.yml")
	raw, err := readRawYAML(mainPath)
	if err != nil {
		t.Fatalf("readRawYAML: %v", err)
	}
	if _, ok := raw["gnbsim"]; ok {
		t.Error("expected gnbsim key to be pruned from main.yml")
	}
	if _, ok := raw["amp"]; ok {
		t.Error("expected amp key to be pruned from main.yml")
	}
	if _, ok := raw["k8s"]; !ok {
		t.Error("expected k8s key in main.yml")
	}
	if _, ok := raw["core"]; !ok {
		t.Error("expected core key in main.yml")
	}
}

func TestComposeConfig_ComponentsSorted(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"5gc", "amp"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	// Components should include 5gc, k8s (implicit), amp.
	comps := out.Body.Components
	sort.Strings(comps)
	if len(comps) != 3 {
		t.Fatalf("expected 3 components (5gc, amp, k8s), got %v", comps)
	}
}
