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
	res, err := s.db.Exec(NodeInsertSQL,
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
	row := s.db.QueryRow(`SELECT `+nodeColumns+` FROM nodes WHERE id = ?`, id)
	return scanNode(row)
}

func (s *Store) UpdateNode(n *Node) error {
	_, err := s.db.Exec(`
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
	res, err := s.db.Exec(`DELETE FROM nodes WHERE id = ?`, id)
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

func (s *Store) DeleteNodesForEntity(entityID string) error {
	_, err := s.db.Exec(`DELETE FROM nodes WHERE entity_id = ?`, entityID)
	return err
}

// ReplaceEntityNodes atomically replaces all nodes for an entity.
func (s *Store) ReplaceEntityNodes(entityID string, nodes []*Node) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM nodes WHERE entity_id = ?`, entityID); err != nil {
		return fmt.Errorf("delete old nodes: %w", err)
	}

	stmt, err := tx.Prepare(NodeInsertSQL)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, n := range nodes {
		n.EntityID = entityID
		res, err := stmt.Exec(n.EntityID, n.ParentID, n.Type, n.Level, n.Seq, n.Content, n.Hash)
		if err != nil {
			return fmt.Errorf("insert node seq %d: %w", n.Seq, err)
		}
		id, _ := res.LastInsertId()
		n.ID = id
	}

	return tx.Commit()
}

func (s *Store) NodeChildren(parentID int64) ([]*Node, error) {
	return s.queryNodes(`
		SELECT `+nodeColumns+` FROM nodes
		WHERE parent_id = ?
		ORDER BY seq`, parentID)
}

// queryNodes executes a query and scans all rows into Node slices.
func (s *Store) queryNodes(query string, args ...any) ([]*Node, error) {
	rows, err := s.db.Query(query, args...)
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
