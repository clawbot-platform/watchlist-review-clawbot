package ofacgraph

import "strings"

type LinkedDocument struct {
    PartyID        string
    DocumentType   string
    DocumentValue  string
    IssuingCountry string
}

func scoreLinkedDocuments(docs []LinkedDocument, normalizedIdentifiers map[string][]string) (int, []string) {
    if len(docs) == 0 || len(normalizedIdentifiers) == 0 {
        return 0, nil
    }

    normalizedDocValues := map[string]struct{}{}
    for _, d := range docs {
        value := strings.ToUpper(strings.TrimSpace(d.DocumentValue))
        if value != "" {
            normalizedDocValues[value] = struct{}{}
        }
    }

    for idType, values := range normalizedIdentifiers {
        for _, value := range values {
            if _, ok := normalizedDocValues[strings.ToUpper(strings.TrimSpace(value))]; ok {
                return 10, []string{"official linked document matches screened identifier: " + strings.TrimSpace(idType)}
            }
        }
    }
    return 0, nil
}
