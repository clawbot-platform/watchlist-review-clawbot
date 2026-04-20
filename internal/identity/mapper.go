package identity

import (
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
)

func BuildCompareRequest(alert *alerts.CanonicalAlert, tenantID string, explain bool) (CompareRequest, error) {
	if alert == nil {
		return CompareRequest{}, fmt.Errorf("alert is required")
	}
	if strings.TrimSpace(alert.ScreenedParty.SourceSystem) == "" || strings.TrimSpace(alert.ScreenedParty.SourceRecordID) == "" {
		return CompareRequest{}, fmt.Errorf("screened_party source reference is required for compare")
	}
	if strings.TrimSpace(alert.MatchedParty.ListSource) == "" || strings.TrimSpace(alert.MatchedParty.ListUID) == "" {
		return CompareRequest{}, fmt.Errorf("matched_party list source and list uid are required for compare")
	}

	return CompareRequest{
		TenantID: tenantID,
		Left: SourceRef{
			SourceSystem:   alert.ScreenedParty.SourceSystem,
			SourceRecordID: alert.ScreenedParty.SourceRecordID,
		},
		Right: SourceRef{
			SourceSystem:   alert.MatchedParty.ListSource,
			SourceRecordID: alert.MatchedParty.ListUID,
		},
		Explain: explain,
	}, nil
}

func BuildScreenOFACRequest(alert *alerts.CanonicalAlert, tenantID string, caseID string) (ScreenOFACRequest, error) {
	if alert == nil {
		return ScreenOFACRequest{}, fmt.Errorf("alert is required")
	}

	subject := OFACSubject{
		Name:        strings.TrimSpace(alert.ScreenedParty.Name.FullName),
		DOB:         strings.TrimSpace(alert.ScreenedParty.DateOfBirth),
		Country:     firstCountry(alert.ScreenedParty),
		Identifiers: identifierMap(alert.ScreenedParty.Identifiers),
		Aliases:     append([]string(nil), alert.ScreenedParty.Name.Aliases...),
		Address:     firstAddress(alert.ScreenedParty),
	}

	if subject.Name == "" {
		return ScreenOFACRequest{}, fmt.Errorf("screened_party.name.full_name is required for screening")
	}

	return ScreenOFACRequest{
		TenantID: tenantID,
		CaseID:   chooseCaseID(caseID, alert.Metadata.CaseID, alert.Metadata.AlertID),
		Subject:  subject,
	}, nil
}

func firstCountry(p alerts.Party) string {
	if len(p.Countries) > 0 && strings.TrimSpace(p.Countries[0]) != "" {
		return strings.TrimSpace(p.Countries[0])
	}
	if len(p.Nationalities) > 0 && strings.TrimSpace(p.Nationalities[0]) != "" {
		return strings.TrimSpace(p.Nationalities[0])
	}
	if len(p.Addresses) > 0 && strings.TrimSpace(p.Addresses[0].Country) != "" {
		return strings.TrimSpace(p.Addresses[0].Country)
	}
	return ""
}

func firstAddress(p alerts.Party) string {
	if len(p.Addresses) == 0 {
		return ""
	}
	addr := p.Addresses[0]
	switch {
	case strings.TrimSpace(addr.AddressText) != "":
		return strings.TrimSpace(addr.AddressText)
	case strings.TrimSpace(addr.City) != "" && strings.TrimSpace(addr.Country) != "":
		return strings.TrimSpace(addr.City) + ", " + strings.TrimSpace(addr.Country)
	case strings.TrimSpace(addr.City) != "":
		return strings.TrimSpace(addr.City)
	case strings.TrimSpace(addr.Country) != "":
		return strings.TrimSpace(addr.Country)
	default:
		return ""
	}
}

func identifierMap(ids []alerts.Identifier) map[string]string {
	if len(ids) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, id := range ids {
		key := strings.TrimSpace(strings.ToLower(string(id.Type)))
		value := strings.TrimSpace(id.Value)
		if key == "" || key == "unknown" || value == "" {
			continue
		}
		if _, exists := out[key]; !exists {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func chooseCaseID(values ...string) string {
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}
