package state

import "time"

// WizardStatus represents the completion status of the setup wizard.
type WizardStatus struct {
	Completed   bool       `json:"completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Steps       []string   `json:"steps,omitempty"`
}

// MetricsSnapshot represents a point-in-time metrics recording.
type MetricsSnapshot struct {
	MetricType string    `json:"metric_type"`
	Data       []byte    `json:"data"`
	RecordedAt time.Time `json:"recorded_at"`
}

// CachedInfo represents cached system information with its collection time.
type CachedInfo struct {
	InfoType    string    `json:"info_type"`
	Data        []byte    `json:"data"`
	CollectedAt time.Time `json:"collected_at"`
}

// Common state keys used in the app_state table.
const (
	KeyWizardCompleted   = "wizard_completed"
	KeyWizardCompletedAt = "wizard_completed_at"
	KeyWizardSteps       = "wizard_steps"
)
