package onramp

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestComposeConfig_ServerEntriesSurviveRoundTrip(t *testing.T) {
	p := newTestProvider(t, baseConfig())

	writeBlueprint(t, p, "main-srsran.yml", map[string]any{
		"srsran": map[string]any{
			"docker": map[string]any{
				"container": map[string]any{
					"gnb_image": "srsran/gnb:latest",
					"ue_image":  "srsran/ue:latest",
				},
				"network": map[string]any{
					"name": "srsran-net",
				},
			},
			"simulation": true,
			"servers": map[string]any{
				"0": map[string]any{
					"gnb_ip":   "",
					"gnb_conf": "deps/srsran/roles/gNB/templates/gnb_zmq.yaml",
					"ue_conf":  "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf",
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

	// Verify the returned typed config has server entries.
	if out.Body.Config.SRSRan == nil {
		t.Fatal("expected srsran section in returned config")
	}
	if len(out.Body.Config.SRSRan.Servers) == 0 {
		t.Fatal("expected srsran.servers in returned config")
	}

	const wantGNBConf = "deps/srsran/roles/gNB/templates/gnb_zmq.yaml"
	const wantUEConf = "deps/srsran/roles/uEsimulator/templates/ue_zmq.conf"

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

	// Verify main.yml on disk can be parsed back into the typed struct.
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
}
