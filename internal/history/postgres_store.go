package history

import (
    "context"
    "database/sql"
    "errors"
    "sort"
    "time"
)

type PostgresStore struct {
    db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
    return &PostgresStore{db: db}
}

func (s *PostgresStore) LookupSamePair(ctx context.Context, req LookupRequest) (*LookupResult, error) {
    if s == nil || s.db == nil {
        return nil, errors.New("history postgres store is not configured")
    }

    samePairRows, err := s.db.QueryContext(ctx, `
        SELECT tenant_id, case_id, alert_id, screened_entity_fingerprint, matched_list_uid,
               matched_program, decision_label, contradiction_pattern, reviewed_at, source
        FROM review_case_dispositions
        WHERE tenant_id = $1
          AND screened_entity_fingerprint = $2
          AND matched_list_uid = $3
        ORDER BY reviewed_at DESC`,
        req.TenantID, req.ScreenedEntityFingerprint, req.MatchedListUID,
    )
    if err != nil {
        return nil, err
    }
    defer samePairRows.Close()

    samePairCases, err := scanCaseDispositions(samePairRows)
    if err != nil {
        return nil, err
    }

    sameProfileRows, err := s.db.QueryContext(ctx, `
        SELECT tenant_id, case_id, alert_id, screened_entity_fingerprint, matched_list_uid,
               matched_program, decision_label, contradiction_pattern, reviewed_at, source
        FROM review_case_dispositions
        WHERE tenant_id = $1
          AND matched_list_uid = $2
        ORDER BY reviewed_at DESC`,
        req.TenantID, req.MatchedListUID,
    )
    if err != nil {
        return nil, err
    }
    defer sameProfileRows.Close()

    sameProfileCases, err := scanCaseDispositions(sameProfileRows)
    if err != nil {
        return nil, err
    }

    res := &LookupResult{
        SamePairCases:           samePairCases,
        SameMatchedProfileCases: sameProfileCases,
    }
    summarizeLookupResult(res)
    return res, nil
}

func (s *PostgresStore) SaveDisposition(ctx context.Context, disposition CaseDisposition) error {
    if s == nil || s.db == nil {
        return errors.New("history postgres store is not configured")
    }
    if disposition.ReviewedAt.IsZero() {
        disposition.ReviewedAt = time.Now().UTC()
    }

    _, err := s.db.ExecContext(ctx, `
        INSERT INTO review_case_dispositions (
            tenant_id, case_id, alert_id, screened_entity_fingerprint, matched_list_uid,
            matched_program, decision_label, contradiction_pattern, reviewed_at, source
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
        ON CONFLICT (tenant_id, case_id) DO UPDATE SET
            alert_id = EXCLUDED.alert_id,
            screened_entity_fingerprint = EXCLUDED.screened_entity_fingerprint,
            matched_list_uid = EXCLUDED.matched_list_uid,
            matched_program = EXCLUDED.matched_program,
            decision_label = EXCLUDED.decision_label,
            contradiction_pattern = EXCLUDED.contradiction_pattern,
            reviewed_at = EXCLUDED.reviewed_at,
            source = EXCLUDED.source`,
        disposition.TenantID,
        disposition.CaseID,
        disposition.AlertID,
        disposition.ScreenedEntityFingerprint,
        disposition.MatchedListUID,
        disposition.MatchedProgram,
        disposition.DecisionLabel,
        disposition.ContradictionPattern,
        disposition.ReviewedAt,
        disposition.Source,
    )
    return err
}

func scanCaseDispositions(rows *sql.Rows) ([]CaseDisposition, error) {
    var out []CaseDisposition
    for rows.Next() {
        var item CaseDisposition
        if err := rows.Scan(
            &item.TenantID,
            &item.CaseID,
            &item.AlertID,
            &item.ScreenedEntityFingerprint,
            &item.MatchedListUID,
            &item.MatchedProgram,
            &item.DecisionLabel,
            &item.ContradictionPattern,
            &item.ReviewedAt,
            &item.Source,
        ); err != nil {
            return nil, err
        }
        out = append(out, item)
    }
    return out, rows.Err()
}

func summarizeLookupResult(res *LookupResult) {
    if res == nil {
        return
    }
    if len(res.SamePairCases) > 0 {
        res.LatestSamePair = &res.SamePairCases[0]
    }

    contradictionSet := map[string]struct{}{}
    var mostRecent *time.Time
    all := append(append([]CaseDisposition{}, res.SamePairCases...), res.SameMatchedProfileCases...)
    sort.Slice(all, func(i, j int) bool { return all[i].ReviewedAt.After(all[j].ReviewedAt) })
    for _, item := range all {
        ts := item.ReviewedAt
        if mostRecent == nil || ts.After(*mostRecent) {
            copyTs := ts
            mostRecent = &copyTs
        }
        switch item.DecisionLabel {
        case DispositionEscalate:
            res.EscalateCount++
        case DispositionInvestigateNextStep:
            res.InvestigateCount++
        case DispositionClose:
            res.CloseCount++
        case DispositionReviewPending:
            res.ReviewPendingCount++
        }
        if item.ContradictionPattern != "" {
            contradictionSet[item.ContradictionPattern] = struct{}{}
        }
    }
    res.MostRecentReviewedAt = mostRecent
    for k := range contradictionSet {
        res.RecurringContradictionKeys = append(res.RecurringContradictionKeys, k)
    }
    sort.Strings(res.RecurringContradictionKeys)
}
