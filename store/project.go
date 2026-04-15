package store

import (
	"database/sql"
	"taskmanager/model"
)

type ProjectStore struct {
	DB *sql.DB
}

func NewProjectStore(db *sql.DB) *ProjectStore {
	return &ProjectStore{DB: db}
}

func (s *ProjectStore) Create(name, description string, ownerId int) (*model.Project, error) {
	result, err := s.DB.Exec(
		"INSERT INTO projects (name, description, owner_id) VALUES (?, ?, ?)",
		name, description, ownerId,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &model.Project{ID: int(id), Name: name, Description: description, OwnerID: ownerId}, nil
}

func (s *ProjectStore) GetAll() ([]model.Project, error) {
	rows, err := s.DB.Query("SELECT id, name, description, owner_id FROM projects")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (s *ProjectStore) GetByID(id int) (*model.Project, error) {
	var p model.Project
	err := s.DB.QueryRow("SELECT id, name, description, owner_id FROM projects WHERE id = ?", id).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
