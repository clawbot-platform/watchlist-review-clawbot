package scoring

type Result struct {
	MatchStrengthScore       int      `json:"match_strength_score"`
	DataSufficiencyScore     int      `json:"data_sufficiency_score"`

	NameMatchScore           int      `json:"name_match_score"`
	DateMatchScore           int      `json:"date_match_score"`
	IdentifierMatchScore     int      `json:"identifier_match_score"`
	GeographyMatchScore      int      `json:"geography_match_score"`
	AddressMatchScore        int      `json:"address_match_score"`
	ContextSupportScore      int      `json:"context_support_score"`

	ScreenedCompletenessScore int     `json:"screened_completeness_score"`
	MatchedCompletenessScore  int     `json:"matched_completeness_score"`
	IdentifierQualityScore    int     `json:"identifier_quality_score"`
	GeographyQualityScore     int     `json:"geography_quality_score"`
	SupportingContextScore    int     `json:"supporting_context_score"`

	Contradictions           []string `json:"contradictions,omitempty"`
	DecisionLabel            string   `json:"decision_label"`
	DecisionReason           string   `json:"decision_reason,omitempty"`
	EvidenceFor              []string `json:"evidence_for,omitempty"`
	EvidenceAgainst          []string `json:"evidence_against,omitempty"`
	MissingInformation       []string `json:"missing_information,omitempty"`
	NextStep                 string   `json:"next_step,omitempty"`
}
