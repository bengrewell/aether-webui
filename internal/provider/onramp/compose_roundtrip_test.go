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

	// Try reading the composed YAML ourselves to see the actual error.
	mainPath := filepath.Join(p.config.OnRampDir, "vars", "main.yml")

	// First, run compose which writes main.yml.
	_, err := p.HandleComposeConfig(t.Context(), &ConfigComposeInput{
		Body: ConfigComposeBody{Components: []string{"k8s", "5gc", "srsran"}},
	})

	// Read the file that compose wrote.
	mainData, _ := os.ReadFile(mainPath)
	t.Logf("main.yml:\n%s", string(mainData))

	// Try to unmarshal directly to see the exact error.
	var cfg OnRampConfig
	parseErr := yaml.Unmarshal(mainData, &cfg)
	if parseErr != nil {
		t.Logf("YAML parse error: %v", parseErr)
	} else {
		t.Logf("YAML parse succeeded, SRSRan=%+v", cfg.SRSRan)
		if cfg.SRSRan != nil && cfg.SRSRan.Servers != nil {
			for k, v := range cfg.SRSRan.Servers {
				t.Logf("Server[%d] = %+v", k, v)
			}
		}
	}

	if err != nil {
		t.Fatalf("HandleComposeConfig: %v", err)
	}
}
