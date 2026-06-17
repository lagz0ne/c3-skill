package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

// SemanticEmbeddingModel names the pinned local sentence-transformer model.
const SemanticEmbeddingModel = "sentence-transformers/all-MiniLM-L6-v2-onnx@1110a243"

const (
	semanticEmbeddingDims = 384
	semanticMinScore      = 0.08
)

// ErrSemanticUnavailable means the local model/runtime cache is not present.
var ErrSemanticUnavailable = errors.New("semantic embedding model unavailable")

// SemanticIndexOptions controls model fetching during index rebuilds.
type SemanticIndexOptions struct {
	AllowDownload bool
}

// SemanticIndexStats describes the work done while making the local semantic
// index match current entity text.
type SemanticIndexStats struct {
	Total   int
	Fresh   int
	Indexed int
	Deleted int
}

// SemanticSearchOptions controls model fetching during query embedding.
type SemanticSearchOptions struct {
	AllowDownload bool
}

type semanticEmbedder interface {
	Embed(ctx context.Context, text string, allowDownload bool) ([]float32, bool, error)
}

var semanticProvider semanticEmbedder = &onnxMiniLMProvider{}

// SetSemanticProviderForTest swaps the process-wide semantic provider.
func SetSemanticProviderForTest(provider semanticEmbedder) func() {
	old := semanticProvider
	semanticProvider = provider
	return func() {
		semanticProvider = old
	}
}

type semanticEmbeddingRow struct {
	entityID string
	hash     string
	vector   []float32
}

type semanticIndexedRow struct {
	model string
	dims  int
	hash  string
}

// RebuildSemanticIndexWithOptions refreshes all semantic vectors. It computes
// vectors before replacing the existing index so failed downloads/runs do not
// erase a previously usable index.
func (s *Store) RebuildSemanticIndexWithOptions(ctx context.Context, opts SemanticIndexOptions) error {
	entities, err := s.AllEntities()
	if err != nil {
		return fmt.Errorf("load semantic entities: %w", err)
	}

	rows := make([]semanticEmbeddingRow, 0, len(entities))
	for _, e := range entities {
		text := s.semanticTextForEntity(e)
		row, ok, err := semanticEmbeddingForText(ctx, e.ID, text, opts.AllowDownload)
		if err != nil {
			return err
		}
		if ok {
			rows = append(rows, row)
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin semantic index rebuild: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM entity_embeddings WHERE model = ?`, SemanticEmbeddingModel); err != nil {
		return fmt.Errorf("clear semantic index: %w", err)
	}
	stmt, err := tx.Prepare(`
		INSERT INTO entity_embeddings(entity_id, model, dims, text_hash, vector, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(entity_id) DO UPDATE SET
			model = excluded.model,
			dims = excluded.dims,
			text_hash = excluded.text_hash,
			vector = excluded.vector,
			updated_at = excluded.updated_at`)
	if err != nil {
		return fmt.Errorf("prepare semantic index insert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.Exec(row.entityID, SemanticEmbeddingModel, semanticEmbeddingDims, row.hash, encodeFloat32Vector(row.vector)); err != nil {
			return fmt.Errorf("insert semantic embedding %s: %w", row.entityID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit semantic index rebuild: %w", err)
	}
	return nil
}

// EnsureSemanticIndexWithOptions makes the semantic index fresh for the current
// entity set. It reuses existing rows when the text hash still matches and only
// embeds missing or stale entities.
func (s *Store) EnsureSemanticIndexWithOptions(ctx context.Context, opts SemanticIndexOptions) (SemanticIndexStats, error) {
	entities, err := s.AllEntities()
	if err != nil {
		return SemanticIndexStats{}, fmt.Errorf("load semantic entities: %w", err)
	}
	indexed, err := s.semanticIndexedRows()
	if err != nil {
		return SemanticIndexStats{}, err
	}

	stats := SemanticIndexStats{Total: len(entities)}
	currentIDs := make(map[string]bool, len(entities))
	type candidate struct {
		entity *Entity
		text   string
	}
	var changed []candidate
	for _, e := range entities {
		text := s.semanticTextForEntity(e)
		hash := semanticTextHash(text)
		currentIDs[e.ID] = true
		row, ok := indexed[e.ID]
		if ok && row.model == SemanticEmbeddingModel && row.dims == semanticEmbeddingDims && row.hash == hash {
			stats.Fresh++
			continue
		}
		changed = append(changed, candidate{entity: e, text: text})
	}

	deleteIDs := make([]string, 0)
	for id := range indexed {
		if !currentIDs[id] {
			deleteIDs = append(deleteIDs, id)
		}
	}

	if len(changed) == 0 && len(deleteIDs) == 0 {
		return stats, nil
	}

	rows := make([]semanticEmbeddingRow, 0, len(changed))
	for _, item := range changed {
		row, ok, err := semanticEmbeddingForText(ctx, item.entity.ID, item.text, opts.AllowDownload)
		if err != nil {
			return stats, err
		}
		if !ok {
			deleteIDs = append(deleteIDs, item.entity.ID)
			continue
		}
		rows = append(rows, row)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return stats, fmt.Errorf("begin semantic index ensure: %w", err)
	}
	defer tx.Rollback()

	for _, id := range deleteIDs {
		res, err := tx.Exec(`DELETE FROM entity_embeddings WHERE entity_id = ?`, id)
		if err != nil {
			return stats, fmt.Errorf("delete stale semantic embedding %s: %w", id, err)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			stats.Deleted += int(n)
		}
	}

	stmt, err := tx.Prepare(`
		INSERT INTO entity_embeddings(entity_id, model, dims, text_hash, vector, updated_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
		ON CONFLICT(entity_id) DO UPDATE SET
			model = excluded.model,
			dims = excluded.dims,
			text_hash = excluded.text_hash,
			vector = excluded.vector,
			updated_at = excluded.updated_at`)
	if err != nil {
		return stats, fmt.Errorf("prepare semantic index ensure insert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.Exec(row.entityID, SemanticEmbeddingModel, semanticEmbeddingDims, row.hash, encodeFloat32Vector(row.vector)); err != nil {
			return stats, fmt.Errorf("ensure semantic embedding %s: %w", row.entityID, err)
		}
		stats.Indexed++
	}
	if err := tx.Commit(); err != nil {
		return stats, fmt.Errorf("commit semantic index ensure: %w", err)
	}
	return stats, nil
}

// SemanticIndexCount returns the number of usable local semantic vectors.
func (s *Store) SemanticIndexCount() (int, error) {
	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM entity_embeddings WHERE model = ? AND dims = ?`,
		SemanticEmbeddingModel, semanticEmbeddingDims,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count semantic index: %w", err)
	}
	return count, nil
}

func (s *Store) semanticIndexedRows() (map[string]semanticIndexedRow, error) {
	rows, err := s.db.Query(`SELECT entity_id, model, dims, text_hash FROM entity_embeddings`)
	if err != nil {
		return nil, fmt.Errorf("load semantic index rows: %w", err)
	}
	defer rows.Close()

	indexed := make(map[string]semanticIndexedRow)
	for rows.Next() {
		var entityID string
		var row semanticIndexedRow
		if err := rows.Scan(&entityID, &row.model, &row.dims, &row.hash); err != nil {
			return nil, fmt.Errorf("scan semantic index row: %w", err)
		}
		indexed[entityID] = row
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan semantic index rows: %w", err)
	}
	return indexed, nil
}

func semanticEmbeddingForText(ctx context.Context, entityID, text string, allowDownload bool) (semanticEmbeddingRow, bool, error) {
	vec, ok, err := embedSemanticText(ctx, text, allowDownload)
	if err != nil {
		return semanticEmbeddingRow{}, false, fmt.Errorf("embed semantic entity %s: %w", entityID, err)
	}
	if !ok {
		return semanticEmbeddingRow{}, false, nil
	}
	return semanticEmbeddingRow{
		entityID: entityID,
		hash:     semanticTextHash(text),
		vector:   vec,
	}, true, nil
}

func semanticTextHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%x", sum[:])
}

func (s *Store) semanticTextForEntity(e *Entity) string {
	var b strings.Builder
	writeSemanticField(&b, "id", e.ID)
	writeSemanticField(&b, "type", e.Type)
	writeSemanticField(&b, "title", e.Title)
	writeSemanticField(&b, "slug", e.Slug)
	writeSemanticField(&b, "category", e.Category)
	writeSemanticField(&b, "goal", e.Goal)
	writeSemanticField(&b, "boundary", e.Boundary)
	writeSemanticField(&b, "parent", e.ParentID)
	if e.Metadata != "" && e.Metadata != "{}" {
		writeSemanticField(&b, "metadata", e.Metadata)
	}
	if nodes, err := s.NodesForEntity(e.ID); err == nil {
		for _, n := range nodes {
			writeSemanticField(&b, n.Type, n.Content)
		}
	}
	if patterns, err := s.CodeMapFor(e.ID); err == nil {
		for _, pattern := range patterns {
			writeSemanticField(&b, "code path", pattern)
		}
	}
	return b.String()
}

func writeSemanticField(b *strings.Builder, label, text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	if label != "" {
		b.WriteString(label)
		b.WriteString(": ")
	}
	b.WriteString(text)
	b.WriteByte('\n')
}

// SearchSemanticWithOptions embeds a query with MiniLM and brute-force scans
// normalized vectors stored in SQLite.
func (s *Store) SearchSemanticWithOptions(ctx context.Context, query, entityType string, limit int, opts SemanticSearchOptions) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	count, err := s.SemanticIndexCount()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}

	qvec, ok, err := embedSemanticText(ctx, query, opts.AllowDownload)
	if err != nil {
		if !opts.AllowDownload {
			return nil, nil
		}
		return nil, fmt.Errorf("embed semantic query: %w", err)
	}
	if !ok {
		return nil, nil
	}

	q := `SELECT e.id, e.type, e.title, e.goal, entity_embeddings.vector
		FROM entity_embeddings
		JOIN entities e ON entity_embeddings.entity_id = e.id
		WHERE entity_embeddings.model = ? AND entity_embeddings.dims = ?`
	args := []any{SemanticEmbeddingModel, semanticEmbeddingDims}
	if entityType != "" {
		q += ` AND e.type = ?`
		args = append(args, entityType)
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var goal sql.NullString
		var blob []byte
		if err := rows.Scan(&r.ID, &r.Type, &r.Title, &goal, &blob); err != nil {
			return nil, fmt.Errorf("semantic search scan: %w", err)
		}
		vec, err := decodeFloat32Vector(blob)
		if err != nil || len(vec) != len(qvec) {
			continue
		}
		score := dotProduct(qvec, vec)
		if score < semanticMinScore {
			continue
		}
		r.Rank = 1 - score
		r.Snippet = strings.TrimSpace(goal.String)
		if r.Snippet == "" {
			r.Snippet = r.Title
		}
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Rank != results[j].Rank {
			return results[i].Rank < results[j].Rank
		}
		return results[i].ID < results[j].ID
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func embedSemanticText(ctx context.Context, text string, allowDownload bool) ([]float32, bool, error) {
	return semanticProvider.Embed(ctx, text, allowDownload)
}

func normalizeVector(vec []float32) {
	var sum float64
	for _, v := range vec {
		sum += float64(v * v)
	}
	if sum == 0 {
		return
	}
	scale := float32(1 / math.Sqrt(sum))
	for i := range vec {
		vec[i] *= scale
	}
}

func dotProduct(a, b []float32) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dot float64
	for i := 0; i < n; i++ {
		dot += float64(a[i] * b[i])
	}
	return dot
}

func encodeFloat32Vector(vec []float32) []byte {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

func decodeFloat32Vector(buf []byte) ([]float32, error) {
	if len(buf)%4 != 0 {
		return nil, fmt.Errorf("invalid vector byte length %d", len(buf))
	}
	vec := make([]float32, len(buf)/4)
	for i := range vec {
		vec[i] = math.Float32frombits(binary.LittleEndian.Uint32(buf[i*4:]))
	}
	return vec, nil
}
