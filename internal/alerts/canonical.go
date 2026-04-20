package alerts

import "strings"

func (a *CanonicalAlert) Normalize() {
	if a == nil {
		return
	}
	a.Kind = AlertKind(strings.TrimSpace(string(a.Kind)))
	a.Metadata.AlertID = strings.TrimSpace(a.Metadata.AlertID)
	a.Metadata.SourceSystem = strings.TrimSpace(strings.ToLower(a.Metadata.SourceSystem))
	a.Metadata.AlertType = strings.TrimSpace(strings.ToLower(a.Metadata.AlertType))
	a.Metadata.Jurisdiction = strings.TrimSpace(strings.ToUpper(a.Metadata.Jurisdiction))
	a.Metadata.CaseID = strings.TrimSpace(a.Metadata.CaseID)
	a.Metadata.Priority = strings.TrimSpace(strings.ToLower(a.Metadata.Priority))

	normalizeParty(&a.ScreenedParty)
	normalizeMatchedParty(&a.MatchedParty)
	normalizeFeatures(&a.ScreeningFeatures)

	if a.Transaction != nil {
		a.Transaction.TransactionID = strings.TrimSpace(a.Transaction.TransactionID)
		a.Transaction.RailType = strings.TrimSpace(strings.ToLower(a.Transaction.RailType))
		a.Transaction.Currency = strings.TrimSpace(strings.ToUpper(a.Transaction.Currency))
		a.Transaction.OriginatorRole = strings.TrimSpace(strings.ToLower(a.Transaction.OriginatorRole))
		a.Transaction.BeneficiaryRole = strings.TrimSpace(strings.ToLower(a.Transaction.BeneficiaryRole))
		a.Transaction.PaymentReference = strings.TrimSpace(a.Transaction.PaymentReference)
		a.Transaction.CountryCorridor = strings.TrimSpace(strings.ToUpper(a.Transaction.CountryCorridor))
		a.Transaction.Institution = strings.TrimSpace(a.Transaction.Institution)
	}

	for i := range a.Artifacts {
		a.Artifacts[i].Kind = strings.TrimSpace(strings.ToLower(a.Artifacts[i].Kind))
		a.Artifacts[i].URI = strings.TrimSpace(a.Artifacts[i].URI)
		a.Artifacts[i].Description = strings.TrimSpace(a.Artifacts[i].Description)
	}
}

func normalizeParty(p *Party) {
	p.EntityType = EntityType(strings.TrimSpace(strings.ToLower(string(p.EntityType))))
	normalizeName(&p.Name)
	p.DateOfBirth = strings.TrimSpace(p.DateOfBirth)
	p.BirthYear = strings.TrimSpace(p.BirthYear)
	p.IncorporationDate = strings.TrimSpace(p.IncorporationDate)
	p.Countries = trimUpperSlice(p.Countries)
	p.Nationalities = trimUpperSlice(p.Nationalities)
	p.SourceRecordID = strings.TrimSpace(p.SourceRecordID)
	p.SourceSystem = strings.TrimSpace(strings.ToLower(p.SourceSystem))
	for i := range p.Addresses {
		normalizeAddress(&p.Addresses[i])
	}
	for i := range p.Identifiers {
		normalizeIdentifier(&p.Identifiers[i])
	}
}

func normalizeMatchedParty(p *MatchedParty) {
	p.ListSource = strings.TrimSpace(strings.ToLower(p.ListSource))
	p.Program = strings.TrimSpace(strings.ToUpper(p.Program))
	p.ListUID = strings.TrimSpace(p.ListUID)
	p.EntityType = EntityType(strings.TrimSpace(strings.ToLower(string(p.EntityType))))
	normalizeName(&p.Name)
	p.DateOfBirth = strings.TrimSpace(p.DateOfBirth)
	p.BirthYear = strings.TrimSpace(p.BirthYear)
	p.IncorporationDate = strings.TrimSpace(p.IncorporationDate)
	p.Countries = trimUpperSlice(p.Countries)
	for i := range p.Addresses {
		normalizeAddress(&p.Addresses[i])
	}
	for i := range p.Identifiers {
		normalizeIdentifier(&p.Identifiers[i])
	}
}

func normalizeFeatures(f *ScreeningFeatures) {
	f.MatchFlags = trimLowerSlice(f.MatchFlags)
	f.ReasonCodes = trimUpperSlice(f.ReasonCodes)
	f.AnalystNotes = trimSpaceSlice(f.AnalystNotes)
	f.ReviewComments = trimSpaceSlice(f.ReviewComments)
}

func normalizeName(n *Name) {
	n.FullName = strings.TrimSpace(n.FullName)
	n.Aliases = trimSpaceSlice(n.Aliases)
	n.NativeName = strings.TrimSpace(n.NativeName)
}

func normalizeAddress(a *Address) {
	a.AddressText = strings.TrimSpace(a.AddressText)
	a.City = strings.TrimSpace(a.City)
	a.Region = strings.TrimSpace(a.Region)
	a.PostalCode = strings.TrimSpace(a.PostalCode)
	a.Country = strings.TrimSpace(strings.ToUpper(a.Country))
}

func normalizeIdentifier(i *Identifier) {
	i.Type = IdentifierType(strings.TrimSpace(strings.ToLower(string(i.Type))))
	i.Value = strings.TrimSpace(i.Value)
	i.IssuingCountry = strings.TrimSpace(strings.ToUpper(i.IssuingCountry))
}

func trimSpaceSlice(in []string) []string {
	var out []string
	for _, v := range in {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func trimUpperSlice(in []string) []string {
	var out []string
	for _, v := range in {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, strings.ToUpper(t))
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func trimLowerSlice(in []string) []string {
	var out []string
	for _, v := range in {
		if t := strings.TrimSpace(v); t != "" {
			out = append(out, strings.ToLower(t))
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
