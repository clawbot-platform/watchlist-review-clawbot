package history

import (
	"fmt"
	"strings"

	"github.com/clawbot-platform/watchlist-review-clawbot/internal/alerts"
	"github.com/clawbot-platform/watchlist-review-clawbot/internal/features"
)

func BuildScreenedEntityFingerprint(alert *alerts.CanonicalAlert, fx *features.ExtractedFeatures) string {
	if alert == nil {
		return ""
	}

	kind := string(alert.Kind)
	name := strings.ToUpper(strings.TrimSpace(alert.ScreenedParty.Name.FullName))
	entityType := strings.ToUpper(string(alert.ScreenedParty.EntityType))
	program := strings.ToUpper(strings.TrimSpace(alert.MatchedParty.Program))

	var identityParts []string
	if fx != nil {
		if fx.Date.ScreenedExact != "" {
			identityParts = append(identityParts, strings.ToUpper(strings.TrimSpace(fx.Date.ScreenedExact)))
		} else if fx.Date.ScreenedYear != "" {
			identityParts = append(identityParts, strings.ToUpper(strings.TrimSpace(fx.Date.ScreenedYear)))
		}

		for _, exact := range fx.Identifiers.ExactMatches {
			if exact.Type != "" || exact.Value != "" {
				part := strings.ToUpper(
					strings.TrimSpace(string(exact.Type)) + ":" + strings.TrimSpace(exact.Value),
				)
				identityParts = append(identityParts, part)
			}
		}
	}

	return fmt.Sprintf("%s|%s|%s|%s|%s", kind, entityType, name, program, strings.Join(identityParts, "|"))
}
