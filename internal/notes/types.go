package notes

type AnalystNote struct {
	Status                    string   `json:"status"`
	Model                     string   `json:"model,omitempty"`
	PromptVersion             string   `json:"prompt_version,omitempty"`
	Note                      string   `json:"note,omitempty"`
	EvidenceSummary           []string `json:"evidence_summary,omitempty"`
	MissingInformationSummary []string `json:"missing_information_summary,omitempty"`
	NextStepRationale         string   `json:"next_step_rationale,omitempty"`
	Warnings                  []string `json:"warnings,omitempty"`
}

const (
	StatusSkipped   = "skipped"
	StatusGenerated = "generated"
	StatusFailed    = "failed"
)
