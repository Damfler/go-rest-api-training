package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"taskmanager/apperror"
	"taskmanager/model"
)

type ProjectRepository struct {
	DB *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{DB: db}
}

func (s *ProjectRepository) Create(ctx context.Context, name, description string, ownerId int) (*model.Project, error) {
	result, err := s.DB.ExecContext(ctx,
		"INSERT INTO projects (name, description, owner_id) VALUES (?, ?, ?)",
		name, description, ownerId,
	)
	if err != nil {
		return nil, fmt.Errorf("ProjectRepository.Create(%s,%s,%d): %w", name, description, ownerId, err)
	}

	id, _ := result.LastInsertId()
	return &model.Project{ID: int(id), Name: name, Description: description, OwnerID: ownerId}, nil
}

func (s *ProjectRepository) GetAll(ctx context.Context) ([]model.Project, error) {
	rows, err := s.DB.QueryContext(ctx, "SELECT id, name, description, owner_id FROM projects")
	if err != nil {
		return nil, fmt.Errorf("ProjectRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("ProjectRepository.GetAll scan: %w", err)
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (s *ProjectRepository) GetByID(ctx context.Context, id int) (*model.Project, error) {
	var p model.Project
	err := s.DB.QueryRowContext(ctx, "SELECT id, name, description, owner_id FROM projects WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperror.NotFoundError{Entity: "project", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("ProjectRepository.GetByID(%d): %w", id, err)
	}
	return &p, nil
}
