package onramp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/danielgtaylor/huma/v2"
	"gopkg.in/yaml.v3"
)

// componentBlueprint maps component names to their blueprint filenames.
// Only components that have a separate blueprint file are included.
var componentBlueprint = map[string]string{
	"srsran":  "main-srsran.yml",
	"ueransim": "main-ueransim.yml",
	"oai":     "main-oai.yml",
	"n3iwf":   "main-n3iwf.yml",
	"sdran":   "main-sdran.yml",
}

// componentYAMLKey maps component names to their top-level YAML key.
var componentYAMLKey = map[string]string{
	"k8s":     "k8s",
	"5gc":     "core",
	"4gc":     "core",
	"gnbsim":  "gnbsim",
	"srsran":  "srsran",
	"ueransim": "ueransim",
	"oai":     "oai",
	"n3iwf":   "n3iwf",
	"sdran":   "sdran",
	"amp":     "amp",
}

// implicitDeps maps a component to its implied dependencies.
var implicitDeps = map[string][]string{
	"5gc": {"k8s"},
	"4gc": {"k8s"},
}

// prunableKeys lists top-level YAML keys that are removed when not selected.
// k8s and core are infrastructure and never pruned.
var prunableKeys = map[string]bool{
	"gnbsim":  true,
	"srsran":  true,
	"ueransim": true,
	"oai":     true,
	"n3iwf":   true,
	"sdran":   true,
	"amp":     true,
}

// HandleComposeConfig builds vars/main.yml from the base config plus selected
// component blueprints, pruning sections for unselected components.
func (o *OnRamp) HandleComposeConfig(_ context.Context, in *ConfigComposeInput) (*ConfigComposeOutput, error) {
	// Validate all components.
	for _, c := range in.Body.Components {
		if _, ok := componentYAMLKey[c]; !ok {
			return nil, huma.Error422UnprocessableEntity(
				fmt.Sprintf("unknown component: %s", c),
				fmt.Errorf("component %q is not in the component registry", c),
			)
		}
	}

	// Expand implicit deps into a deduplicated, sorted slice for
	// deterministic merge order and stable output.
	seen := make(map[string]bool, len(in.Body.Components))
	for _, c := range in.Body.Components {
		seen[c] = true
		for _, dep := range implicitDeps[c] {
			seen[dep] = true
		}
	}
	components := make([]string, 0, len(seen))
	for c := range seen {
		components = append(components, c)
	}
	sort.Strings(components)

	varsDir := filepath.Join(o.config.OnRampDir, "vars")
	mainPath := filepath.Join(varsDir, "main.yml")

	// Read base config as untyped map.
	base, err := readRawYAML(mainPath)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to read base config", err)
	}

	// Merge blueprints in sorted order for deterministic results.
	var activeBlueprints []string
	for _, comp := range components {
		bpFile, hasBP := componentBlueprint[comp]
		if !hasBP {
			continue
		}
		bpPath := filepath.Join(varsDir, bpFile)
		bp, err := readRawYAML(bpPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, huma.Error404NotFound(
					fmt.Sprintf("blueprint %s not found", bpFile),
					fmt.Errorf("blueprint file %s does not exist", bpPath),
				)
			}
			return nil, huma.Error500InternalServerError("failed to read blueprint", err)
		}

		yamlKey := componentYAMLKey[comp]

		// Merge the component's top-level key from the blueprint.
		if bpSection, ok := bp[yamlKey]; ok {
			if bpMap, ok := bpSection.(map[string]any); ok {
				if baseMap, ok := base[yamlKey].(map[string]any); ok {
					deepMergeMaps(baseMap, bpMap)
				} else {
					base[yamlKey] = bpSection
				}
			} else {
				base[yamlKey] = bpSection
			}
		}

		// Merge the blueprint's core section into the base core.
		if bpCore, ok := bp["core"]; ok {
			if bpCoreMap, ok := bpCore.(map[string]any); ok {
				if baseCore, ok := base["core"].(map[string]any); ok {
					deepMergeMaps(baseCore, bpCoreMap)
				} else {
					base["core"] = bpCore
				}
			}
		}

		activeBlueprints = append(activeBlueprints, bpFile)
	}

	// Determine which YAML keys are "kept" by selected components.
	keptKeys := make(map[string]bool)
	for _, comp := range components {
		if key, ok := componentYAMLKey[comp]; ok {
			keptKeys[key] = true
		}
	}

	// Prune unselected prunable keys.
	for key := range prunableKeys {
		if !keptKeys[key] {
			delete(base, key)
		}
	}

	// Marshal and write.
	data, err := yaml.Marshal(base)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to marshal config", err)
	}
	if err := os.WriteFile(mainPath, data, 0o644); err != nil {
		return nil, huma.Error500InternalServerError("failed to write config", err)
	}

	// Re-read as typed config.
	cfg, err := o.readVarsFile(mainPath)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to re-read config", err)
	}

	return &ConfigComposeOutput{
		Body: ConfigComposeResult{
			ActiveBlueprints: activeBlueprints,
			Components:       components,
			Config:           cfg,
		},
	}, nil
}

// readRawYAML reads a YAML file into an untyped map.
func readRawYAML(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}
