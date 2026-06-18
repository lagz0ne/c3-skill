package store

import (
	"database/sql"
	"fmt"
)

// Node represents an element in the content tree.
type Node struct {
	ID       int64
	EntityID string
	ParentID sql.NullInt64
	Type     string
	Level    int
	Seq      int
	Content  string
	Hash     string
}

const nodeColumns = `id, entity_id, parent_id, type, level, seq, content, hash`

const NodeInsertSQL = `INSERT INTO nodes (entity_id, parent_id, type, level, seq, content, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

func (s *Store) InsertNode(n *Node) (int64, error) {
	res, err := s.exec.Exec(NodeInsertSQL,
		n.EntityID, n.ParentID, n.Type, n.Level, n.Seq, n.Content, n.Hash,
	)
	if err != nil {
		return 0, fmt.Errorf("insert node: %w", err)
	}
	id, _ := res.LastInsertId()
	n.ID = id
	return id, nil
}

func (s *Store) GetNode(id int64) (*Node, error) {
	row := s.exec.QueryRow(`SELECT `+nodeColumns+` FROM nodes WHERE id = ?`, id)
	return scanNode(row)
}

func (s *Store) UpdateNode(n *Node) error {
	_, err := s.exec.Exec(`
		UPDATE nodes SET type = ?, level = ?, seq = ?, content = ?, hash = ?,
			parent_id = ?
		WHERE id = ?`,
		n.Type, n.Level, n.Seq, n.Content, n.Hash, n.ParentID, n.ID,
	)
	if err != nil {
		return fmt.Errorf("update node %d: %w", n.ID, err)
	}
	return nil
}

// DeleteNode removes a node. Children are cascade-deleted by FK.
func (s *Store) DeleteNode(id int64) error {
	res, err := s.exec.Exec(`DELETE FROM nodes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete node %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("node %d not found", id)
	}
	return nil
}

// NodesForEntity returns all nodes for an entity, ordered for tree reconstruction.
func (s *Store) NodesForEntity(entityID string) ([]*Node, error) {
	return s.queryNodes(`
		SELECT `+nodeColumns+` FROM nodes
		WHERE entity_id = ?
		ORDER BY parent_id NULLS FIRST, seq`, entityID)
}

// ReplaceEntityNodes atomically replaces all nodes for an entity, preserving each
// node's existing ParentID. Standalone it opens its own transaction; inside an
// apply it enlists in the open one.
func (s *Store) ReplaceEntityNodes(entityID string, nodes []*Node) error {
	return s.WithTx(func(ts *Store) error {
		if _, err := ts.exec.Exec(`DELETE FROM nodes WHERE entity_id = ?`, entityID); err != nil {
			return fmt.Errorf("delete old nodes: %w", err)
		}
		for _, n := range nodes {
			n.EntityID = entityID
			res, err := ts.exec.Exec(NodeInsertSQL, n.EntityID, n.ParentID, n.Type, n.Level, n.Seq, n.Content, n.Hash)
			if err != nil {
				return fmt.Errorf("insert node seq %d: %w", n.Seq, err)
			}
			id, _ := res.LastInsertId()
			n.ID = id
		}
		return nil
	})
}

// AppendNodeTree appends a parsed node subtree after an entity's existing nodes:
// each snippet root (parentIndex[i] < 0) becomes a new root node sequenced after the
// current last root, and the rest keep their snippet parent. Existing nodes are left
// untouched (their content, hashes, and order are preserved — only new nodes are
// added), so the entity reseals to a new merkle without drifting any prior block.
// Standalone it opens its own transaction; inside an apply it enlists in the open one.
func (s *Store) AppendNodeTree(entityID string, nodes []*Node, parentIndex []int) error {
	return s.WithTx(func(ts *Store) error {
		var maxRootSeq int
		if err := ts.exec.QueryRow(
			`SELECT COALESCE(MAX(seq), -1) FROM nodes WHERE entity_id = ? AND parent_id IS NULL`,
			entityID,
		).Scan(&maxRootSeq); err != nil {
			return fmt.Errorf("append node tree: max root seq: %w", err)
		}
		realIDs := make([]int64, len(nodes))
		rootCount := 0
		for i, n := range nodes {
			n.EntityID = entityID
			if parentIndex[i] < 0 {
				n.ParentID = sql.NullInt64{} // a new root node
				n.Seq = maxRootSeq + 1 + rootCount
				rootCount++
			} else {
				n.ParentID = sql.NullInt64{Int64: realIDs[parentIndex[i]], Valid: true}
			}
			res, err := ts.exec.Exec(NodeInsertSQL, n.EntityID, n.ParentID, n.Type, n.Level, n.Seq, n.Content, n.Hash)
			if err != nil {
				return fmt.Errorf("append node %d: %w", i, err)
			}
			id, _ := res.LastInsertId()
			n.ID = id
			realIDs[i] = id
		}
		return nil
	})
}

// InsertNodeTree replaces an entity's nodes from a parsed tree, remapping each
// node's tree-index parent (parentIndex[i] < 0 ⇒ root) to the real autoincrement
// ID assigned on insert. Standalone it opens its own transaction; inside an apply
// it enlists in the open one — so a created fact's body lands atomically with its
// entity row, version, and seal.
func (s *Store) InsertNodeTree(entityID string, nodes []*Node, parentIndex []int) error {
	return s.WithTx(func(ts *Store) error {
		if _, err := ts.exec.Exec(`DELETE FROM nodes WHERE entity_id = ?`, entityID); err != nil {
			return fmt.Errorf("delete old nodes: %w", err)
		}
		realIDs := make([]int64, len(nodes))
		for i, n := range nodes {
			n.EntityID = entityID
			if pi := parentIndex[i]; pi >= 0 {
				n.ParentID = sql.NullInt64{Int64: realIDs[pi], Valid: true}
			}
			res, err := ts.exec.Exec(NodeInsertSQL, n.EntityID, n.ParentID, n.Type, n.Level, n.Seq, n.Content, n.Hash)
			if err != nil {
				return fmt.Errorf("insert node %d: %w", i, err)
			}
			id, _ := res.LastInsertId()
			n.ID = id
			realIDs[i] = id
		}
		return nil
	})
}

func (s *Store) NodeChildren(parentID int64) ([]*Node, error) {
	return s.queryNodes(`
		SELECT `+nodeColumns+` FROM nodes
		WHERE parent_id = ?
		ORDER BY seq`, parentID)
}

// queryNodes executes a query and scans all rows into Node slices.
func (s *Store) queryNodes(query string, args ...any) ([]*Node, error) {
	rows, err := s.exec.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Node
	for rows.Next() {
		n, err := scanNodeFromRows(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func scanNode(row *sql.Row) (*Node, error) {
	var n Node
	err := row.Scan(&n.ID, &n.EntityID, &n.ParentID, &n.Type, &n.Level, &n.Seq, &n.Content, &n.Hash)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func scanNodeFromRows(rows *sql.Rows) (*Node, error) {
	var n Node
	err := rows.Scan(&n.ID, &n.EntityID, &n.ParentID, &n.Type, &n.Level, &n.Seq, &n.Content, &n.Hash)
	if err != nil {
		return nil, err
	}
	return &n, nil
}
