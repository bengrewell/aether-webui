package onramp

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// TestComposeConfig_RawYAMLBlueprint_AllFieldsSurvive verifies that compose
// with a raw YAML blueprint (bare integer keys, like the real OnRamp files)
// preserves all srsran fields: docker, simulation, gnb_conf, ue_conf.
func TestComposeConfig_RawYAMLBlueprint_AllFieldsSurvive(t *testing.T) {
	p := newTestProvider(t, baseConfig())
	writeRawBlueprint(t, p, "main-srsran.yml", rawBlueprintSRSRan)

	out, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "srsran"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	// --- Verify returned typed config ---
	cfg := out.Body.Config
	if cfg.SRSRan == nil {
		t.Fatal("returned config: srsran section is nil")
	}

	// Docker
	if cfg.SRSRan.Docker == nil {
		t.Fatal("returned config: srsran.docker is nil")
	}
	if cfg.SRSRan.Docker.Container == nil {
		t.Fatal("returned config: srsran.docker.container is nil")
	}
	if got := cfg.SRSRan.Docker.Container.GNBImage; got != "srsran/gnb:latest" {
		t.Errorf("returned docker.container.gnb_image = %q, want %q", got, "srsran/gnb:latest")
	}
	if got := cfg.SRSRan.Docker.Container.UEImage; got != "srsran/ue:latest" {
		t.Errorf("returned docker.container.ue_image = %q, want %q", got, "srsran/ue:latest")
	}
	if cfg.SRSRan.Docker.Network == nil {
		t.Fatal("returned config: srsran.docker.network is nil")
	}
	if got := cfg.SRSRan.Docker.Network.Name; got != "srsran-net" {
		t.Errorf("returned docker.network.name = %q, want %q", got, "srsran-net")
	}

	// Simulation
	if cfg.SRSRan.Simulation == nil {
		t.Fatal("returned config: srsran.simulation is nil")
	}
	if !*cfg.SRSRan.Simulation {
		t.Error("returned config: srsran.simulation = false, want true")
	}

	// Servers
	if len(cfg.SRSRan.Servers) == 0 {
		t.Fatal("returned config: srsran.servers is empty")
	}
	srv, ok := cfg.SRSRan.Servers[0]
	if !ok {
		t.Fatal("returned config: srsran.servers[0] missing")
	}
	const wantGNBConf = "deps/srsran/roles/gNB/templates/gnb_zmq.yaml"
	const wantUEConf = "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf"
	if srv.GNBConf != wantGNBConf {
		t.Errorf("returned servers[0].gnb_conf = %q, want %q", srv.GNBConf, wantGNBConf)
	}
	if srv.UEConf != wantUEConf {
		t.Errorf("returned servers[0].ue_conf = %q, want %q", srv.UEConf, wantUEConf)
	}

	// --- Verify disk round-trip via readVarsFile (production read path) ---
	mainPath := filepath.Join(p.config.OnRampDir, "vars", "main.yml")
	diskCfg, err := p.readVarsFile(mainPath)
	if err != nil {
		t.Fatalf("readVarsFile main.yml: %v", err)
	}

	if diskCfg.SRSRan == nil {
		t.Fatal("disk: srsran section is nil")
	}
	if diskCfg.SRSRan.Docker == nil {
		t.Fatal("disk: srsran.docker is nil")
	}
	if diskCfg.SRSRan.Docker.Container == nil {
		t.Fatal("disk: srsran.docker.container is nil")
	}
	if diskCfg.SRSRan.Docker.Container.GNBImage != "srsran/gnb:latest" {
		t.Errorf("disk docker.container.gnb_image = %q", diskCfg.SRSRan.Docker.Container.GNBImage)
	}
	if diskCfg.SRSRan.Simulation == nil || !*diskCfg.SRSRan.Simulation {
		t.Error("disk: srsran.simulation should be true")
	}
	diskSrv, ok := diskCfg.SRSRan.Servers[0]
	if !ok {
		t.Fatal("disk: srsran.servers[0] missing")
	}
	if diskSrv.GNBConf != wantGNBConf {
		t.Errorf("disk servers[0].gnb_conf = %q, want %q", diskSrv.GNBConf, wantGNBConf)
	}
	if diskSrv.UEConf != wantUEConf {
		t.Errorf("disk servers[0].ue_conf = %q, want %q", diskSrv.UEConf, wantUEConf)
	}
}

// TestComposeConfig_ThenApplyDefaults_PreservesFields simulates the full
// compose → apply-defaults flow: compose writes the config, then a partial
// patch (like apply-defaults setting gnb_ip) is merged via deepMergeConfig.
// All non-patched fields must survive.
func TestComposeConfig_ThenApplyDefaults_PreservesFields(t *testing.T) {
	p := newTestProvider(t, baseConfig())
	writeRawBlueprint(t, p, "main-srsran.yml", rawBlueprintSRSRan)

	// Step 1: compose
	_, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "srsran"}},
	})
	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}

	// Step 2: read the composed config back (as apply-defaults would)
	mainPath := filepath.Join(p.config.OnRampDir, "vars", "main.yml")
	cfg, err := p.readVarsFile(mainPath)
	if err != nil {
		t.Fatalf("readVarsFile: %v", err)
	}

	// Step 3: simulate apply-defaults patch — sets gnb_ip
	patchJSON := []byte(`{"srsran":{"servers":{"0":{"gnb_ip":"10.103.102.30"}}}}`)
	merged, err := deepMergeConfig(&cfg, patchJSON)
	if err != nil {
		t.Fatalf("deepMergeConfig: %v", err)
	}

	// Step 4: write merged config back
	if err := p.writeVarsFile(mainPath, &merged); err != nil {
		t.Fatalf("writeVarsFile: %v", err)
	}

	// Step 5: read final result and verify ALL fields survived
	final, err := p.readVarsFile(mainPath)
	if err != nil {
		t.Fatalf("readVarsFile (final): %v", err)
	}

	if final.SRSRan == nil {
		t.Fatal("final: srsran is nil")
	}

	// gnb_ip should be patched
	srv, ok := final.SRSRan.Servers[0]
	if !ok {
		t.Fatal("final: srsran.servers[0] missing")
	}
	if srv.GNBIP != "10.103.102.30" {
		t.Errorf("final servers[0].gnb_ip = %q, want %q", srv.GNBIP, "10.103.102.30")
	}

	// gnb_conf and ue_conf must still be present
	const wantGNBConf = "deps/srsran/roles/gNB/templates/gnb_zmq.yaml"
	const wantUEConf = "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf"
	if srv.GNBConf != wantGNBConf {
		t.Errorf("final servers[0].gnb_conf = %q, want %q", srv.GNBConf, wantGNBConf)
	}
	if srv.UEConf != wantUEConf {
		t.Errorf("final servers[0].ue_conf = %q, want %q", srv.UEConf, wantUEConf)
	}

	// Docker must survive
	if final.SRSRan.Docker == nil {
		t.Fatal("final: srsran.docker is nil")
	}
	if final.SRSRan.Docker.Container == nil {
		t.Fatal("final: srsran.docker.container is nil")
	}
	if final.SRSRan.Docker.Container.GNBImage != "srsran/gnb:latest" {
		t.Errorf("final docker.container.gnb_image = %q", final.SRSRan.Docker.Container.GNBImage)
	}
	if final.SRSRan.Docker.Network == nil || final.SRSRan.Docker.Network.Name != "srsran-net" {
		t.Errorf("final docker.network.name missing or wrong")
	}

	// Simulation must survive
	if final.SRSRan.Simulation == nil || !*final.SRSRan.Simulation {
		t.Error("final: srsran.simulation should be true")
	}
}

// TestDeepMergeConfig_DoesNotDropSiblingKeys verifies that deepMergeConfig
// preserves sibling keys when patching a nested leaf.
func TestDeepMergeConfig_DoesNotDropSiblingKeys(t *testing.T) {
	sim := true
	base := OnRampConfig{
		SRSRan: &SRSRanConfig{
			Docker: &SRSRanDocker{
				Container: &SRSRanContainer{
					GNBImage: "srsran/gnb:latest",
					UEImage:  "srsran/ue:latest",
				},
				Network: &SRSRanNetwork{Name: "srsran-net"},
			},
			Simulation: &sim,
			Servers: map[int]*SRSRanServer{
				0: {
					GNBIP:   "",
					GNBConf: "deps/srsran/roles/gNB/templates/gnb_zmq.yaml",
					UEConf:  "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf",
				},
			},
		},
	}

	patch := `{"srsran":{"servers":{"0":{"gnb_ip":"10.1.2.3"}}}}`
	merged, err := deepMergeConfig(&base, []byte(patch))
	if err != nil {
		t.Fatalf("deepMergeConfig: %v", err)
	}

	if merged.SRSRan == nil {
		t.Fatal("merged srsran is nil")
	}
	if merged.SRSRan.Docker == nil {
		t.Fatal("merged srsran.docker lost")
	}
	if merged.SRSRan.Docker.Container == nil || merged.SRSRan.Docker.Container.GNBImage != "srsran/gnb:latest" {
		t.Error("merged docker.container.gnb_image lost")
	}
	if merged.SRSRan.Simulation == nil || !*merged.SRSRan.Simulation {
		t.Error("merged simulation lost")
	}

	srv := merged.SRSRan.Servers[0]
	if srv == nil {
		t.Fatal("merged servers[0] nil")
	}
	if srv.GNBIP != "10.1.2.3" {
		t.Errorf("gnb_ip = %q, want %q", srv.GNBIP, "10.1.2.3")
	}
	if srv.GNBConf != "deps/srsran/roles/gNB/templates/gnb_zmq.yaml" {
		t.Errorf("gnb_conf = %q, want preserved", srv.GNBConf)
	}
	if srv.UEConf != "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf" {
		t.Errorf("ue_conf = %q, want preserved", srv.UEConf)
	}
}

// TestConvertKeysToStrings_IntegerKeys verifies that convertKeysToStrings
// handles map[any]any with integer keys (produced by yaml.v3 for bare 0: keys).
func TestConvertKeysToStrings_IntegerKeys(t *testing.T) {
	// Simulate what yaml.v3 produces for bare integer keys.
	input := map[string]any{
		"srsran": map[string]any{
			"servers": map[any]any{
				0: map[string]any{
					"gnb_ip":   "",
					"gnb_conf": "path/to/gnb.yaml",
					"ue_conf":  "path/to/ue.conf",
				},
			},
			"docker": map[string]any{
				"container": map[string]any{
					"gnb_image": "srsran/gnb:latest",
				},
			},
			"simulation": true,
		},
	}

	result := convertKeysToStrings(input)
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var cfg OnRampConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if cfg.SRSRan == nil {
		t.Fatal("srsran nil after conversion")
	}
	if cfg.SRSRan.Docker == nil || cfg.SRSRan.Docker.Container == nil {
		t.Fatal("docker lost after conversion")
	}
	if cfg.SRSRan.Docker.Container.GNBImage != "srsran/gnb:latest" {
		t.Errorf("gnb_image = %q", cfg.SRSRan.Docker.Container.GNBImage)
	}
	if cfg.SRSRan.Simulation == nil || !*cfg.SRSRan.Simulation {
		t.Error("simulation lost after conversion")
	}
	srv, ok := cfg.SRSRan.Servers[0]
	if !ok || srv == nil {
		t.Fatal("servers[0] lost after conversion")
	}
	if srv.GNBConf != "path/to/gnb.yaml" {
		t.Errorf("gnb_conf = %q", srv.GNBConf)
	}
}
