package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

var knownMetadataKeys = map[string]bool{
	"c3-version":  true,
	"c3-seal":     true,
	"summary":     true,
	"description": true,
}

type canonicalDoc struct {
	Path          string
	ID            string
	Title         string
	Type          string
	Category      string
	ParentID      string
	Goal          string
	Boundary      string
	Status        string
	Date          string
	Body          string
	C3Version     any
	Summary       any
	Description   any
	Seal          string
	Relationships map[string][]string
	Extra         map[string]any
}

func parseMetadataMap(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil || metadata == nil {
		return map[string]any{}
	}
	return metadata
}

func marshalMetadataMap(metadata map[string]any) string {
	if len(metadata) == 0 {
		return "{}"
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func buildMetadataFromFrontmatter(summary, description string, extra map[string]interface{}) string {
	metadata := make(map[string]any, len(extra)+2)
	for key, value := range extra {
		metadata[key] = value
	}
	if summary != "" {
		metadata["summary"] = summary
	}
	if description != "" {
		metadata["description"] = description
	}
	return marshalMetadataMap(metadata)
}

func canonicalDateText(raw string) string {
	t, ok := parseCanonicalDate(raw)
	if !ok {
		return raw
	}
	return t.Format("2006-01-02")
}

func canonicalDateSlug(raw string) string {
	t, ok := parseCanonicalDate(raw)
	if !ok {
		return raw
	}
	return t.Format("20060102")
}

func parseCanonicalDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
		"20060102",
	} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func writeYAMLField(b *strings.Builder, key string, value any) {
	payload, err := yaml.Marshal(map[string]any{key: value})
	if err != nil {
		fmt.Fprintf(b, "%s: %v\n", key, value)
		return
	}
	b.Write(payload)
}

func metadataKeys(metadata map[string]any) []string {
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func normalizeStringSlice(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	sort.Strings(out)
	return out
}

func copyMetadataExcludingKnown(metadata map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range metadata {
		if knownMetadataKeys[key] {
			continue
		}
		out[key] = value
	}
	return out
}

func computeCanonicalSeal(doc canonicalDoc) string {
	sealed := doc
	sealed.Seal = ""
	sum := sha256.Sum256([]byte(renderCanonicalDoc(sealed, false)))
	return fmt.Sprintf("%x", sum)
}

func canonicalBody(body string) string {
	return strings.Trim(body, "\n")
}

func optionalString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func renderCanonicalDoc(doc canonicalDoc, includeSeal bool) string {
	var b strings.Builder

	b.WriteString("---\n")
	writeYAMLField(&b, "id", doc.ID)
	if doc.C3Version != nil {
		writeYAMLField(&b, "c3-version", doc.C3Version)
	}
	if includeSeal {
		seal := doc.Seal
		if seal == "" {
			seal = computeCanonicalSeal(doc)
		}
		writeYAMLField(&b, "c3-seal", seal)
	}
	writeYAMLField(&b, "title", doc.Title)
	if doc.Type != "" && doc.Type != "system" {
		writeYAMLField(&b, "type", doc.Type)
	}
	if doc.Category != "" {
		writeYAMLField(&b, "category", doc.Category)
	}
	if doc.Boundary != "" {
		writeYAMLField(&b, "boundary", doc.Boundary)
	}
	if doc.ParentID != "" {
		writeYAMLField(&b, "parent", doc.ParentID)
	}
	if doc.Goal != "" {
		writeYAMLField(&b, "goal", doc.Goal)
	}
	if doc.Summary != nil {
		writeYAMLField(&b, "summary", doc.Summary)
	}
	if doc.Status != "" && doc.Status != "active" {
		writeYAMLField(&b, "status", doc.Status)
	}
	if doc.Date != "" {
		writeYAMLField(&b, "date", canonicalDateText(doc.Date))
	}
	if doc.Description != nil {
		writeYAMLField(&b, "description", doc.Description)
	}
	for _, relType := range []string{"uses", "affects", "scope", "sources", "origin"} {
		if ids, ok := doc.Relationships[relType]; ok && len(ids) > 0 {
			writeYAMLField(&b, relType, normalizeStringSlice(ids))
		}
	}
	for _, key := range metadataKeys(doc.Extra) {
		writeYAMLField(&b, key, doc.Extra[key])
	}

	b.WriteString("---\n")
	body := canonicalBody(doc.Body)
	if strings.TrimSpace(body) != "" {
		b.WriteString("\n")
		b.WriteString(body)
		b.WriteString("\n")
	}
	return b.String()
}

func canonicalDocFromParsedDoc(doc frontmatter.ParsedDoc) canonicalDoc {
	fm := doc.Frontmatter
	rels := map[string][]string{
		"uses":    append([]string{}, fm.Refs...),
		"affects": append([]string{}, fm.Affects...),
		"scope":   append([]string{}, fm.Scope...),
		"sources": append([]string{}, fm.Sources...),
		"origin":  append([]string{}, fm.Origin...),
	}
	extra := map[string]any{}
	for key, value := range fm.Extra {
		extra[key] = value
	}
	docType := fm.Type
	if docType == "" && fm.ID == "c3-0" {
		docType = "system"
	}
	return canonicalDoc{
		Path:          doc.Path,
		ID:            fm.ID,
		Title:         fm.Title,
		Type:          docType,
		Category:      fm.Category,
		ParentID:      fm.Parent,
		Goal:          fm.Goal,
		Boundary:      fm.Boundary,
		Status:        fm.Status,
		Date:          fm.Date,
		Body:          doc.Body,
		C3Version:     fm.Extra["c3-version"],
		Summary:       optionalString(fm.Summary),
		Description:   optionalString(fm.Description),
		Seal:          fm.Seal,
		Relationships: rels,
		Extra:         copyMetadataExcludingKnown(extra),
	}
}

func verifyParsedDocSeal(doc frontmatter.ParsedDoc) (string, string) {
	canonical := canonicalDocFromParsedDoc(doc)
	expected := computeCanonicalSeal(canonical)
	return canonical.Seal, expected
}
