package wizard

import "time"

const wizardNamespace = "_wizard"

// validSteps defines the recognized wizard step names.
var validSteps = map[string]bool{
	"nodes":      true,
	"preflight":  true,
	"roles":      true,
	"config":     true,
	"deployment": true,
}

// stepData is the payload stored in the objects table for each completed step.
type stepData struct {
	CompletedAt time.Time `json:"completed_at"`
}

// StepStatus represents the completion state of a single wizard step.
type StepStatus struct {
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at"`
}

// ActiveTask describes a currently running onramp task, included in the wizard
// state so the frontend can resume monitoring after a page refresh.
type ActiveTask struct {
	ID        string `json:"id"`
	Component string `json:"component"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Status    string `json:"status"`
}

// WizardState is the full wizard status returned by GET /api/v1/wizard.
type WizardState struct {
	Completed  bool                  `json:"completed"`
	Steps      map[string]StepStatus `json:"steps"`
	ActiveTask *ActiveTask           `json:"active_task"`
}

// ---------------------------------------------------------------------------
// Huma I/O types
// ---------------------------------------------------------------------------

type WizardGetOutput struct {
	Body WizardState
}

type StepCompleteInput struct {
	Step string `path:"step" doc:"Wizard step name"`
}

type StepCompleteOutput struct {
	Body StepStatus
}

type WizardResetOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}
