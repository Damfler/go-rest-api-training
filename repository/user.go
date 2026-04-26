// repository/user.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"taskmanager/apperror"
	"taskmanager/model"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(ctx context.Context, name, email string) (*model.User, error) {
	result, err := r.DB.ExecContext(ctx,
		"INSERT INTO users (name, email) VALUES (?, ?)",
		name, email,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, &apperror.ConflictError{
				Entity: "user", Field: "email", Value: email,
			}
		}
		return nil, fmt.Errorf("UserRepository.Create: %w", err)
	}

	id, _ := result.LastInsertId()
	return &model.User{ID: int(id), Name: name, Email: email}, nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT id, name, email FROM users")
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetAll: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, fmt.Errorf("UserRepository.GetAll scan: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	var u model.User
	err := r.DB.QueryRowContext(ctx,
		"SELECT id, name, email FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Name, &u.Email)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperror.NotFoundError{Entity: "user", ID: id}
	}
	if err != nil {
		return nil, fmt.Errorf("UserRepository.GetByID: %w", err)
	}

	return &u, nil
}
