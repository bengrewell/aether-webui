package wizard

import (
	"context"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/bengrewell/aether-webui/internal/store"
)

func (w *Wizard) handleGet(ctx context.Context, _ *struct{}) (*WizardGetOutput, error) {
	keys, err := w.Store().List(ctx, wizardNamespace)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list wizard state", err)
	}

	completed := make(map[string]stepData)
	for _, k := range keys {
		item, ok, err := store.Load[stepData](w.Store(), ctx, k)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to load wizard step", err)
		}
		if ok {
			completed[k.ID] = item.Data
		}
	}

	steps := make(map[string]StepStatus, len(validSteps))
	allDone := true
	for name := range validSteps {
		if data, ok := completed[name]; ok {
			t := data.CompletedAt
			steps[name] = StepStatus{Completed: true, CompletedAt: &t}
		} else {
			steps[name] = StepStatus{Completed: false}
			allDone = false
		}
	}

	return &WizardGetOutput{Body: WizardState{
		Completed: allDone && len(completed) == len(validSteps),
		Steps:     steps,
	}}, nil
}

func (w *Wizard) handleCompleteStep(ctx context.Context, in *StepCompleteInput) (*StepCompleteOutput, error) {
	if !validSteps[in.Step] {
		return nil, huma.Error422UnprocessableEntity(
			fmt.Sprintf("invalid step %q", in.Step),
		)
	}

	now := time.Now().UTC()
	key := store.Key{Namespace: wizardNamespace, ID: in.Step}
	if _, err := store.Save(w.Store(), ctx, key, stepData{CompletedAt: now}); err != nil {
		return nil, huma.Error500InternalServerError("failed to save wizard step", err)
	}

	return &StepCompleteOutput{Body: StepStatus{
		Completed:   true,
		CompletedAt: &now,
	}}, nil
}

func (w *Wizard) handleReset(ctx context.Context, _ *struct{}) (*WizardResetOutput, error) {
	keys, err := w.Store().List(ctx, wizardNamespace)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list wizard state", err)
	}
	for _, k := range keys {
		if err := w.Store().Delete(ctx, k); err != nil {
			return nil, huma.Error500InternalServerError("failed to delete wizard step", err)
		}
	}
	out := &WizardResetOutput{}
	out.Body.Message = "wizard state reset"
	return out, nil
}
