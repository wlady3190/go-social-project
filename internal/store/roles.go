package store

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string  `json:"name"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

type RolesStore struct {
	db *sql.DB
}

func (s *RolesStore) GetByName(ctx context.Context, slug string) (*Role, error) {
	query := `
	SELECT id, name, description, level FROM roles where id = $1
	`
	role := &Role{}

	err := s.db.QueryRowContext(ctx, query, slug).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.Level,
	)

	if err != nil {
		return nil, err
	}

	return role, nil

}
