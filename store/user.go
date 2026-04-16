package store

import (
	"database/sql"
	"taskmanager/apperror"
	"taskmanager/model"
)

type UserStore struct {
	DB *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{DB: db}
}

func (s *UserStore) Create(name, email string) (*model.User, error) {
	result, err := s.DB.Exec(
		"INSERT INTO users (name, email) VALUES (?, ?)",
		name, email,
	)
	if err != nil {
		return nil, &apperror.ConflictError{
			Entity: "user",
			Field:  "email",
			Value:  email,
		}
	}

	id, _ := result.LastInsertId()
	return &model.User{ID: int(id), Name: name, Email: email}, nil
}

func (s *UserStore) GetAll() ([]model.User, error) {
	rows, err := s.DB.Query("SELECT id, name, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Name, &u.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (s *UserStore) GetByID(id int) (*model.User, error) {
	var u model.User
	err := s.DB.QueryRow(
		"SELECT id, name, email FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Name, &u.Email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
