package store

import (
	"encoding/json"
	"fmt"
)

// EvalMatchRecord is the latest local matched-state result for one eval spec.
// It is disposable cache data: enough to compare same/different across runs
// without storing verbatim gathered material by default.
type EvalMatchRecord struct {
	Fact          string
	Claim         string
	FactRoot      string
	EvalSpecHash  string
	ExternalState string
	Verdict       string
	Evidence      []string
	Units         []EvalMatchUnit
	CacheUnits    []EvalMatchUnit
}

// EvalMatchUnit is one refined matched surface after gather/transform.
type EvalMatchUnit struct {
	Kind   string
	Key    string
	Digest string
	Bytes  int
}

// SaveEvalMatch replaces the latest matched-state record for a fact.
func (s *Store) SaveEvalMatch(record EvalMatchRecord) error {
	if record.Fact == "" {
		return fmt.Errorf("eval match: missing fact")
	}
	evidence, err := json.Marshal(record.Evidence)
	if err != nil {
		return fmt.Errorf("eval match evidence: %w", err)
	}
	return s.WithTx(func(ts *Store) error {
		if _, err := ts.exec.Exec(`
			INSERT INTO eval_matches(fact, claim, fact_root, eval_spec_hash, external_state, verdict, evidence, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))
			ON CONFLICT(fact) DO UPDATE SET
				claim = excluded.claim,
				fact_root = excluded.fact_root,
				eval_spec_hash = excluded.eval_spec_hash,
				external_state = excluded.external_state,
				verdict = excluded.verdict,
				evidence = excluded.evidence,
				updated_at = datetime('now')`,
			record.Fact, record.Claim, record.FactRoot, record.EvalSpecHash,
			record.ExternalState, record.Verdict, string(evidence),
		); err != nil {
			return fmt.Errorf("save eval match: %w", err)
		}
		if _, err := ts.exec.Exec(`DELETE FROM eval_match_units WHERE fact = ?`, record.Fact); err != nil {
			return fmt.Errorf("replace eval match units: %w", err)
		}
		if err := ts.saveEvalMatchUnits("eval_match_units", record.Fact, record.Units); err != nil {
			return err
		}
		if _, err := ts.exec.Exec(`DELETE FROM eval_match_cache_units WHERE fact = ?`, record.Fact); err != nil {
			return fmt.Errorf("replace eval match cache units: %w", err)
		}
		cacheUnits := record.CacheUnits
		if len(cacheUnits) == 0 {
			cacheUnits = record.Units
		}
		if err := ts.saveEvalMatchUnits("eval_match_cache_units", record.Fact, cacheUnits); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) saveEvalMatchUnits(table, fact string, units []EvalMatchUnit) error {
	for i, unit := range units {
		if _, err := s.exec.Exec(fmt.Sprintf(`
			INSERT INTO %s(fact, unit_index, kind, key, digest, bytes)
			VALUES (?, ?, ?, ?, ?, ?)`, table),
			fact, i, unit.Kind, unit.Key, unit.Digest, unit.Bytes,
		); err != nil {
			return fmt.Errorf("insert %s unit %d: %w", table, i, err)
		}
	}
	return nil
}

// EvalMatch returns the latest matched-state record for a fact.
func (s *Store) EvalMatch(fact string) (EvalMatchRecord, error) {
	var record EvalMatchRecord
	var evidence string
	err := s.exec.QueryRow(`
		SELECT fact, claim, fact_root, eval_spec_hash, external_state, verdict, evidence
		FROM eval_matches WHERE fact = ?`, fact).
		Scan(&record.Fact, &record.Claim, &record.FactRoot, &record.EvalSpecHash, &record.ExternalState, &record.Verdict, &evidence)
	if err != nil {
		return EvalMatchRecord{}, err
	}
	if err := json.Unmarshal([]byte(evidence), &record.Evidence); err != nil {
		return EvalMatchRecord{}, fmt.Errorf("eval match evidence: %w", err)
	}
	record.Units, err = s.evalMatchUnits("eval_match_units", fact)
	if err != nil {
		return EvalMatchRecord{}, err
	}
	record.CacheUnits, err = s.evalMatchUnits("eval_match_cache_units", fact)
	if err != nil {
		return EvalMatchRecord{}, err
	}
	return record, nil
}

func (s *Store) evalMatchUnits(table, fact string) ([]EvalMatchUnit, error) {
	rows, err := s.exec.Query(fmt.Sprintf(`
		SELECT kind, key, digest, bytes
		FROM %s WHERE fact = ?
		ORDER BY unit_index`, table), fact)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var units []EvalMatchUnit
	for rows.Next() {
		var unit EvalMatchUnit
		if err := rows.Scan(&unit.Kind, &unit.Key, &unit.Digest, &unit.Bytes); err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return units, nil
}
