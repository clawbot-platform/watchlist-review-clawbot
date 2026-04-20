package alerts

import (
	"fmt"
	"strings"
)

type FieldError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	if e == nil || len(e.Errors) == 0 {
		return ""
	}
	parts := make([]string, 0, len(e.Errors))
	for _, fe := range e.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s", fe.Field, fe.Message))
	}
	return strings.Join(parts, "; ")
}

func (e *ValidationError) add(field, msg string) { e.Errors = append(e.Errors, FieldError{Field: field, Message: msg}) }
func (e *ValidationError) hasErrors() bool       { return e != nil && len(e.Errors) > 0 }

func (a *CanonicalAlert) Validate() error {
	if a == nil {
		return &ValidationError{Errors: []FieldError{{Field: "alert", Message: "is required"}}}
	}
	a.Normalize()
	verr := &ValidationError{}
	if a.Kind == "" || a.Kind == AlertKindUnknown { verr.add("kind", "must be set to a supported alert kind") }
	if strings.TrimSpace(a.Metadata.AlertID) == "" { verr.add("metadata.alert_id", "is required") }
	if strings.TrimSpace(a.Metadata.SourceSystem) == "" { verr.add("metadata.source_system", "is required") }
	if a.Metadata.CreatedAt.IsZero() { verr.add("metadata.created_at", "is required") }
	validateParty(verr, "screened_party", a.ScreenedParty)
	validateMatchedParty(verr, "matched_party", a.MatchedParty)

	switch a.Kind {
	case AlertKindIndividualOnboarding:
		if a.ScreenedParty.EntityType != EntityTypeIndividual {
			verr.add("screened_party.entity_type", "must be individual for individual_onboarding")
		}
		if a.Transaction != nil {
			verr.add("transaction", "must be omitted for onboarding alerts")
		}
	case AlertKindOrganizationOnboarding:
		if a.ScreenedParty.EntityType != EntityTypeOrganization {
			verr.add("screened_party.entity_type", "must be organization for organization_onboarding")
		}
		if a.Transaction != nil {
			verr.add("transaction", "must be omitted for onboarding alerts")
		}
	case AlertKindACHParty:
		if a.Transaction == nil {
			verr.add("transaction", "is required for ach_party alerts")
		} else {
			validateTransaction(verr, "transaction", *a.Transaction)
		}
	}
	if verr.hasErrors() {
		return verr
	}
	return nil
}

func validateParty(verr *ValidationError, prefix string, party Party) {
	if party.EntityType == "" || party.EntityType == EntityTypeUnknown {
		verr.add(prefix+".entity_type", "is required")
	}
	if strings.TrimSpace(party.Name.FullName) == "" {
		verr.add(prefix+".name.full_name", "is required")
	}
}

func validateMatchedParty(verr *ValidationError, prefix string, party MatchedParty) {
	if strings.TrimSpace(party.ListSource) == "" {
		verr.add(prefix+".list_source", "is required")
	}
	if party.EntityType == "" || party.EntityType == EntityTypeUnknown {
		verr.add(prefix+".entity_type", "is required")
	}
	if strings.TrimSpace(party.Name.FullName) == "" {
		verr.add(prefix+".name.full_name", "is required")
	}
}

func validateTransaction(verr *ValidationError, prefix string, tx TransactionContext) {
	if strings.TrimSpace(tx.TransactionID) == "" {
		verr.add(prefix+".transaction_id", "is required")
	}
	if strings.TrimSpace(tx.RailType) == "" {
		verr.add(prefix+".rail_type", "is required")
	}
}
