package wizard

import "time"

const wizardNamespace = "_wizard"

// validSteps defines the recognized wizard step names.
var validSteps = map[string]bool{
	"nodes":      true,
	"preflight":  true,
	"roles":      true,
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

// WizardState is the full wizard status returned by GET /api/v1/wizard.
type WizardState struct {
	Completed bool                  `json:"completed"`
	Steps     map[string]StepStatus `json:"steps"`
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
