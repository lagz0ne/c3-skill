package walker

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
)

// C3Entity represents a single entity in the C3 architecture graph.
type C3Entity struct {
	ID            string
	Type          frontmatter.DocType
	Title         string
	Slug          string
	Path          string // relative to .c3/
	Frontmatter   *frontmatter.Frontmatter
	Body          string
	Relationships []string // IDs this entity references
}

// C3Graph is an in-memory graph of C3 entities with relationship queries.
type C3Graph struct {
	entities map[string]*C3Entity
	byType   map[frontmatter.DocType][]*C3Entity
}

// ParseWarning records a .md file that has YAML frontmatter delimiters but failed to parse.
type ParseWarning struct {
	Path string // relative to .c3/
}

// WalkResult holds both successfully parsed docs and files that failed to parse.
type WalkResult struct {
	Docs     []frontmatter.ParsedDoc
	Warnings []ParseWarning
}

// WalkC3Docs recursively walks a .c3/ directory and parses all .md files.
func WalkC3Docs(c3Dir string) ([]frontmatter.ParsedDoc, error) {
	result, err := WalkC3DocsWithWarnings(c3Dir)
	if err != nil {
		return nil, err
	}
	return result.Docs, nil
}

// WalkC3DocsWithWarnings is like WalkC3Docs but also returns parse warnings.
func WalkC3DocsWithWarnings(c3Dir string) (*WalkResult, error) {
	result := &WalkResult{}

	err := filepath.Walk(c3Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && info.Name() == "_index" {
			return filepath.SkipDir
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		raw := string(content)
		fm, body := frontmatter.ParseFrontmatter(raw)
		rel, _ := filepath.Rel(c3Dir, path)
		if fm != nil {
			result.Docs = append(result.Docs, frontmatter.ParsedDoc{
				Frontmatter: fm,
				Body:        body,
				Path:        rel,
			})
		} else if strings.HasPrefix(raw, "---\n") {
			result.Warnings = append(result.Warnings, ParseWarning{Path: rel})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

var slugPattern = regexp.MustCompile(`^(c3-\d+-|c3-\d+|ref-|recipe-|adr-\d+-|README)`)

// SlugFromPath derives a slug from a file path by stripping the ID prefix.
// For README.md files (containers), the slug is derived from the parent directory name.
func SlugFromPath(filePath string) string {
	base := strings.TrimSuffix(filepath.Base(filePath), ".md")
	if base == "README" {
		dir := filepath.Dir(filePath)
		if dir == "." || dir == "" || dir == "/" {
			return "" // top-level context README
		}
		dirBase := filepath.Base(dir)
		return slugPattern.ReplaceAllString(dirBase, "")
	}
	return slugPattern.ReplaceAllString(base, "")
}

// BuildGraph constructs a C3Graph from parsed documents.
func BuildGraph(docs []frontmatter.ParsedDoc) *C3Graph {
	g := &C3Graph{
		entities: make(map[string]*C3Entity),
		byType:   make(map[frontmatter.DocType][]*C3Entity),
	}

	for _, doc := range docs {
		docType := frontmatter.ClassifyDoc(doc.Frontmatter)
		if docType == frontmatter.DocUnknown {
			continue
		}

		title := doc.Frontmatter.Title
		if title == "" {
			title = doc.Frontmatter.ID
		}

		entity := &C3Entity{
			ID:            doc.Frontmatter.ID,
			Type:          docType,
			Title:         title,
			Slug:          SlugFromPath(doc.Path),
			Path:          doc.Path,
			Frontmatter:   doc.Frontmatter,
			Body:          doc.Body,
			Relationships: frontmatter.DeriveRelationships(doc.Frontmatter),
		}

		g.entities[entity.ID] = entity
		g.byType[docType] = append(g.byType[docType], entity)
	}

	return g
}

// Len returns the number of entities in the graph.
func (g *C3Graph) Len() int {
	return len(g.entities)
}

// Get returns an entity by ID, or nil if not found.
func (g *C3Graph) Get(id string) *C3Entity {
	return g.entities[id]
}

// All returns all entities.
func (g *C3Graph) All() []*C3Entity {
	result := make([]*C3Entity, 0, len(g.entities))
	for _, e := range g.entities {
		result = append(result, e)
	}
	return result
}

// ByType returns all entities of the given type.
func (g *C3Graph) ByType(t frontmatter.DocType) []*C3Entity {
	return g.byType[t]
}

// Children returns entities whose parent field equals parentId.
func (g *C3Graph) Children(parentId string) []*C3Entity {
	var result []*C3Entity
	for _, e := range g.entities {
		if e.Frontmatter.Parent == parentId {
			result = append(result, e)
		}
	}
	return result
}

// RefsFor returns ref entities cited by the given entity's refs field.
func (g *C3Graph) RefsFor(entityId string) []*C3Entity {
	entity := g.entities[entityId]
	if entity == nil {
		return nil
	}
	var result []*C3Entity
	for _, refId := range entity.Frontmatter.Refs {
		if ref := g.entities[refId]; ref != nil {
			result = append(result, ref)
		}
	}
	return result
}

// CitedBy returns entities that cite the given refId in their refs or scope fields.
func (g *C3Graph) CitedBy(refId string) []*C3Entity {
	var result []*C3Entity
	for _, e := range g.entities {
		if contains(e.Frontmatter.Refs, refId) || contains(e.Frontmatter.Scope, refId) {
			result = append(result, e)
		}
	}
	return result
}

// Forward returns entities that this entity directly affects:
// children, entities in affects list, and (for refs) entities that cite this ref.
func (g *C3Graph) Forward(id string) []*C3Entity {
	entity := g.entities[id]
	if entity == nil {
		return nil
	}

	var result []*C3Entity
	result = append(result, g.Children(id)...)

	for _, affectedId := range entity.Frontmatter.Affects {
		if affected := g.entities[affectedId]; affected != nil {
			result = append(result, affected)
		}
	}

	if entity.Type == frontmatter.DocRef {
		result = append(result, g.CitedBy(id)...)
	}

	return result
}

// Reverse returns entities that point to the given id via relationships, parent, or affects.
func (g *C3Graph) Reverse(id string) []*C3Entity {
	var result []*C3Entity
	for _, e := range g.entities {
		if contains(e.Relationships, id) || e.Frontmatter.Parent == id || contains(e.Frontmatter.Affects, id) {
			result = append(result, e)
		}
	}
	return result
}

// Transitive performs BFS from id up to the given depth, returning all reachable entities.
func (g *C3Graph) Transitive(id string, depth int) []*C3Entity {
	visited := map[string]bool{id: true}
	var result []*C3Entity
	frontier := []string{id}

	for d := 0; d < depth && len(frontier) > 0; d++ {
		var nextFrontier []string
		for _, currentId := range frontier {
			for _, entity := range g.Forward(currentId) {
				if !visited[entity.ID] {
					visited[entity.ID] = true
					result = append(result, entity)
					nextFrontier = append(nextFrontier, entity.ID)
				}
			}
		}
		frontier = nextFrontier
	}

	return result
}

func contains(s []string, v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}
