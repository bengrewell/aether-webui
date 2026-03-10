package configdefaults

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"gopkg.in/yaml.v3"

	"github.com/bengrewell/aether-webui/internal/nodefacts"
	"github.com/bengrewell/aether-webui/internal/provider/onramp"
	"github.com/bengrewell/aether-webui/internal/store"
)

const (
	factsNamespace = "_nodefacts"
	factsTTL       = 5 * time.Minute
)

// handleGetNodeFacts returns discovered network facts for a single node.
func (p *Provider) handleGetNodeFacts(ctx context.Context, in *NodeFactsGetInput) (*NodeFactsGetOutput, error) {
	st := p.Store()

	node, ok, err := st.GetNode(ctx, in.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to get node", err)
	}
	if !ok {
		return nil, huma.Error404NotFound("node not found", fmt.Errorf("no node with id %s", in.ID))
	}

	facts, err := p.getOrGatherFacts(ctx, st, node, in.Refresh)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to gather facts", err)
	}

	return &NodeFactsGetOutput{Body: facts}, nil
}

// handleApplyConfigDefaults gathers facts for all nodes, computes defaults,
// and merges them into vars/main.yml.
func (p *Provider) handleApplyConfigDefaults(ctx context.Context, in *ConfigDefaultsApplyInput) (*ConfigDefaultsApplyOutput, error) {
	st := p.Store()
	log := p.Log()

	// List all nodes.
	nodeInfos, err := st.ListNodes(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list nodes", err)
	}

	if len(nodeInfos) == 0 {
		result := ConfigDefaultsResult{
			Applied: []AppliedDefault{},
			Errors:  []string{"no nodes registered"},
		}
		// Return current config even with no nodes.
		cfg, readErr := p.readVarsFile()
		if readErr == nil {
			result.Config = cfg
		}
		return &ConfigDefaultsApplyOutput{Body: result}, nil
	}

	// Gather facts for each node.
	var (
		factsMap = make(map[string]nodefacts.NodeFacts)
		errs     []string
	)
	for _, ni := range nodeInfos {
		// Fetch full node (with credentials) for SSH.
		node, ok, err := st.GetNode(ctx, ni.ID)
		if err != nil || !ok {
			errs = append(errs, fmt.Sprintf("node %s: could not load", ni.Name))
			continue
		}

		facts, err := p.getOrGatherFacts(ctx, st, node, in.Refresh)
		if err != nil {
			errs = append(errs, fmt.Sprintf("node %s: %v", ni.Name, err))
			continue
		}
		if facts.Error != "" {
			errs = append(errs, fmt.Sprintf("node %s: %s", ni.Name, facts.Error))
		}
		facts.NodeID = node.ID
		facts.NodeName = node.Name
		factsMap[node.ID] = facts
	}

	// Evaluate rules against node facts.
	var applied []AppliedDefault
	for _, ni := range nodeInfos {
		facts, ok := factsMap[ni.ID]
		if !ok {
			continue
		}
		for _, rule := range defaultRules {
			if !matchesRole(ni.Roles, rule.Roles) {
				continue
			}
			val, explanation := rule.ComputeFn(facts)
			if val == nil {
				continue
			}
			applied = append(applied, AppliedDefault{
				Field:       rule.Field,
				Value:       val,
				Explanation: explanation,
				SourceNode:  ni.Name,
			})
		}
	}

	// Read current config, apply defaults, write back.
	mainYML := filepath.Join(p.onRampDir, "vars", "main.yml")
	cfg, err := p.readVarsFile()
	if err != nil {
		log.Error("failed to read vars/main.yml", "error", err)
		return nil, huma.Error500InternalServerError("failed to read config", err)
	}

	if len(applied) > 0 {
		patchMap := buildPatchMap(applied)
		patchJSON, err := json.Marshal(patchMap)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to build defaults patch", err)
		}

		cfg, err = deepMergeConfig(&cfg, patchJSON)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to merge defaults", err)
		}

		if err := writeVarsFile(mainYML, &cfg); err != nil {
			return nil, huma.Error500InternalServerError("failed to write config", err)
		}
	}

	return &ConfigDefaultsApplyOutput{Body: ConfigDefaultsResult{
		Applied: applied,
		Errors:  errs,
		Config:  cfg,
	}}, nil
}

// getOrGatherFacts returns cached facts if valid, or gathers new ones via SSH.
func (p *Provider) getOrGatherFacts(ctx context.Context, st store.Client, node store.Node, forceRefresh bool) (nodefacts.NodeFacts, error) {
	key := store.Key{Namespace: factsNamespace, ID: node.ID}

	if !forceRefresh {
		item, ok, err := store.Load[nodefacts.NodeFacts](st, ctx, key)
		if err == nil && ok {
			// Check if cache is still valid: not expired and node hasn't been updated since.
			if item.Data.GatheredAt.After(node.UpdatedAt) {
				return item.Data, nil
			}
		}
	}

	// Gather fresh facts via SSH.
	facts, err := p.gatherer.Gather(ctx, node.AnsibleHost, node.AnsibleUser, string(node.Password), node.SSHKey)
	if err != nil {
		return nodefacts.NodeFacts{}, err
	}
	facts.NodeID = node.ID
	facts.NodeName = node.Name

	// Cache the result.
	if _, err := store.Save(st, ctx, key, facts, store.WithTTL(factsTTL)); err != nil {
		p.Log().Warn("failed to cache node facts", "node_id", node.ID, "error", err)
	}

	return facts, nil
}

// buildPatchMap converts applied defaults into a nested map suitable for
// JSON marshaling and deep-merging into OnRampConfig.
// e.g. "core.amf.ip" → {"core": {"amf": {"ip": "10.0.0.10"}}}
func buildPatchMap(applied []AppliedDefault) map[string]any {
	root := make(map[string]any)
	for _, a := range applied {
		parts := strings.Split(a.Field, ".")
		m := root
		for i, part := range parts {
			if i == len(parts)-1 {
				m[part] = a.Value
			} else {
				if _, ok := m[part]; !ok {
					m[part] = make(map[string]any)
				}
				m = m[part].(map[string]any)
			}
		}
	}
	return root
}

// --- YAML/config helpers (duplicated from onramp to avoid import cycle) ---

func (p *Provider) readVarsFile() (onramp.OnRampConfig, error) {
	path := filepath.Join(p.onRampDir, "vars", "main.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return onramp.OnRampConfig{}, err
	}
	var cfg onramp.OnRampConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return onramp.OnRampConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

func writeVarsFile(path string, cfg *onramp.OnRampConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func deepMergeConfig(base *onramp.OnRampConfig, patchJSON []byte) (onramp.OnRampConfig, error) {
	baseData, err := json.Marshal(base)
	if err != nil {
		return onramp.OnRampConfig{}, fmt.Errorf("marshal base: %w", err)
	}
	var baseMap map[string]any
	if err := json.Unmarshal(baseData, &baseMap); err != nil {
		return onramp.OnRampConfig{}, fmt.Errorf("unmarshal base: %w", err)
	}

	var patchMap map[string]any
	if err := json.Unmarshal(patchJSON, &patchMap); err != nil {
		return onramp.OnRampConfig{}, fmt.Errorf("unmarshal patch: %w", err)
	}

	deepMergeMaps(baseMap, patchMap)

	merged, err := json.Marshal(baseMap)
	if err != nil {
		return onramp.OnRampConfig{}, fmt.Errorf("marshal merged: %w", err)
	}
	var result onramp.OnRampConfig
	if err := json.Unmarshal(merged, &result); err != nil {
		return onramp.OnRampConfig{}, fmt.Errorf("unmarshal merged: %w", err)
	}
	return result, nil
}

func deepMergeMaps(base, patch map[string]any) {
	for k, pv := range patch {
		bv, exists := base[k]
		if !exists {
			base[k] = pv
			continue
		}
		bm, bok := bv.(map[string]any)
		pm, pok := pv.(map[string]any)
		if bok && pok {
			deepMergeMaps(bm, pm)
		} else {
			base[k] = pv
		}
	}
}
