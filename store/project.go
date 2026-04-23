package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"taskmanager/apperror"
	"taskmanager/model"
)

type ProjectStore struct {
	DB *sql.DB
}

func NewProjectStore(db *sql.DB) *ProjectStore {
	return &ProjectStore{DB: db}
}

func (s *ProjectStore) Create(ctx context.Context, name, description string, ownerId int) (*model.Project, error) {
	result, err := s.DB.ExecContext(ctx,
		"INSERT INTO projects (name, description, owner_id) VALUES (?, ?, ?)",
		name, description, ownerId,
	)
	if err != nil {
		return nil, fmt.Errorf("ProjectStore.Create(%s,%s,%d): %w", name, description, ownerId, err)
	}

	id, _ := result.LastInsertId()
	return &model.Project{ID: int(id), Name: name, Description: description, OwnerID: ownerId}, nil
}

func (s *ProjectStore) GetAll(ctx context.Context) ([]model.Project, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id, name, description, owner_id FROM projects")
	if err != nil {
		return nil, fmt.Errorf("ProjectStore.GetAll: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("ProjectStore.GetAll scan: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (s *ProjectStore) GetByID(ctx context.Context, id int) (*model.Project, error) {
	var p model.Project
	err := s.DB.QueryRowContext(ctx, "SELECT id, name, description, owner_id FROM projects WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperror.NotFoundError{Entity: "project", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("ProjectStore.GetByID(%d): %w", id, err)
	}
	return &p, nil
}
