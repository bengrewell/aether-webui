package preflight

import (
	"context"
	"fmt"
	"sync"

	"github.com/danielgtaylor/huma/v2"
)

func (p *Preflight) handleListChecks(ctx context.Context, _ *struct{}) (*PreflightListOutput, error) {
	deps := DefaultDeps(p.Store(), p.Log())
	results := make([]CheckResult, len(registry))

	var wg sync.WaitGroup
	wg.Add(len(registry))
	for i, check := range registry {
		go func(idx int, c Check) {
			defer wg.Done()
			results[idx] = c.RunCheck(ctx, deps)
		}(i, check)
	}
	wg.Wait()

	passed, failed := 0, 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	return &PreflightListOutput{
		Body: PreflightSummary{
			Passed:  passed,
			Failed:  failed,
			Total:   len(results),
			Results: results,
		},
	}, nil
}

func (p *Preflight) handleGetCheck(ctx context.Context, in *PreflightGetInput) (*PreflightGetOutput, error) {
	idx, ok := checkIndex[in.ID]
	if !ok {
		return nil, huma.Error404NotFound(fmt.Sprintf("check %q not found", in.ID))
	}

	deps := DefaultDeps(p.Store(), p.Log())
	result := registry[idx].RunCheck(ctx, deps)
	return &PreflightGetOutput{Body: result}, nil
}

func (p *Preflight) handleFixCheck(ctx context.Context, in *PreflightFixInput) (*PreflightFixOutput, error) {
	idx, ok := checkIndex[in.ID]
	if !ok {
		return nil, huma.Error404NotFound(fmt.Sprintf("check %q not found", in.ID))
	}

	check := registry[idx]
	if check.RunFix == nil {
		return nil, huma.Error422UnprocessableEntity(fmt.Sprintf("check %q has no automated fix", in.ID))
	}

	deps := DefaultDeps(p.Store(), p.Log())
	result := check.RunFix(ctx, deps)
	return &PreflightFixOutput{Body: result}, nil
}
