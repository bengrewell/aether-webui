package onramp

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestComposeConfig_ServerEntriesSurviveRoundTrip uses raw YAML (bare integer
// keys) to match the real OnRamp blueprint format and verifies all srsran
// fields survive the compose → disk round-trip.
func TestComposeConfig_ServerEntriesSurviveRoundTrip(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	// Write blueprint as raw YAML with bare integer keys, matching the real
	// OnRamp file format. Go map literals use string keys ("0":) which
	// produce different yaml.v3 internal types than bare YAML integers (0:).
	writeRawBlueprint(t, p, "main-srsran.yml", rawBlueprintSRSRan)

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "srsran"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	// Verify the returned typed config has all srsran fields.
	if out.Body.Config.SRSRan == nil {
		t.Fatal("expected srsran section in returned config")
	}

	const wantGNBConf = "deps/srsran/roles/gNB/templates/gnb_zmq.yaml"
	const wantUEConf = "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf"

	// Servers
	if len(out.Body.Config.SRSRan.Servers) == 0 {
		t.Fatal("expected srsran.servers in returned config")
	}
	retSrv, ok := out.Body.Config.SRSRan.Servers[0]
	if !ok {
		t.Fatal("expected srsran.servers[0] in returned config")
	}
	if retSrv.GNBConf != wantGNBConf {
		t.Errorf("returned gnb_conf = %q, want %q", retSrv.GNBConf, wantGNBConf)
	}
	if retSrv.UEConf != wantUEConf {
		t.Errorf("returned ue_conf = %q, want %q", retSrv.UEConf, wantUEConf)
	}

	// Docker
	if out.Body.Config.SRSRan.Docker == nil {
		t.Fatal("expected srsran.docker in returned config")
	}
	if out.Body.Config.SRSRan.Docker.Container == nil {
		t.Fatal("expected srsran.docker.container in returned config")
	}
	if got := out.Body.Config.SRSRan.Docker.Container.GNBImage; got != "srsran/gnb:latest" {
		t.Errorf("returned docker.container.gnb_image = %q, want %q", got, "srsran/gnb:latest")
	}
	if got := out.Body.Config.SRSRan.Docker.Container.UEImage; got != "srsran/ue:latest" {
		t.Errorf("returned docker.container.ue_image = %q, want %q", got, "srsran/ue:latest")
	}
	if out.Body.Config.SRSRan.Docker.Network == nil || out.Body.Config.SRSRan.Docker.Network.Name != "srsran-net" {
		t.Error("expected srsran.docker.network.name = srsran-net")
	}

	// Simulation
	if out.Body.Config.SRSRan.Simulation == nil || !*out.Body.Config.SRSRan.Simulation {
		t.Error("expected srsran.simulation = true in returned config")
	}

	// Verify main.yml on disk can be parsed back with all fields intact.
	mainPath := filepath.Join(p.config.OnRampDir, "vars", "main.yml")
	mainData, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("ReadFile main.yml: %v", err)
	}

	var cfg OnRampConfig
	if err := yaml.Unmarshal(mainData, &cfg); err != nil {
		t.Fatalf("YAML unmarshal main.yml: %v", err)
	}

	if cfg.SRSRan == nil {
		t.Fatal("expected srsran section on disk")
	}

	// Servers on disk
	if cfg.SRSRan.Servers == nil {
		t.Fatal("expected srsran.servers on disk")
	}
	diskSrv, ok := cfg.SRSRan.Servers[0]
	if !ok {
		t.Fatal("expected srsran.servers[0] on disk")
	}
	if diskSrv.GNBConf != wantGNBConf {
		t.Errorf("disk gnb_conf = %q, want %q", diskSrv.GNBConf, wantGNBConf)
	}
	if diskSrv.UEConf != wantUEConf {
		t.Errorf("disk ue_conf = %q, want %q", diskSrv.UEConf, wantUEConf)
	}

	// Docker on disk
	if cfg.SRSRan.Docker == nil || cfg.SRSRan.Docker.Container == nil {
		t.Fatal("expected srsran.docker.container on disk")
	}
	if cfg.SRSRan.Docker.Container.GNBImage != "srsran/gnb:latest" {
		t.Errorf("disk docker.container.gnb_image = %q", cfg.SRSRan.Docker.Container.GNBImage)
	}

	// Simulation on disk
	if cfg.SRSRan.Simulation == nil || !*cfg.SRSRan.Simulation {
		t.Error("expected srsran.simulation = true on disk")
	}
}
