package service

import (
	"context"
	"errors"
	"taskmanager/apperror"
	"taskmanager/model"
	"testing"
)

type mockUserRepo struct {
	users []model.User
	err   error
}

func (m *mockUserRepo) Create(ctx context.Context, name, email string) (*model.User, error) {
	if m.err != nil {
		return nil, m.err
	}

	for _, u := range m.users {
		if u.Email == email {
			return nil, &apperror.ConflictError{
				Entity: "user", Field: "email", Value: email,
			}
		}
	}

	user := &model.User{ID: len(m.users) + 1, Name: name, Email: email}
	m.users = append(m.users, *user)
	return user, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]model.User, error) {
	return m.users, m.err
}

func TestUserServiceCreate(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewUserService(repo)

	user, err := svc.Create(context.Background(), model.CreateUserRequest{
		Name:  "Alex",
		Email: "alex@mail.com",
	})

	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "Alex" {
		t.Errorf("name = %q, want Alex", user.Name)
	}
}

func TestUserServiceValidation(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewUserService(repo)

	_, err := svc.Create(context.Background(), model.CreateUserRequest{
		Name:  "",
		Email: "invalid",
	})

	if err == nil {
		t.Error("expected validation error")
	}

	// Repository не должен быть вызван при ошибке валидации
	if len(repo.users) != 0 {
		t.Error("repo should not be called on validation error")
	}
}

func TestUserServiceDuplicateEmail(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewUserService(repo)

	_, err := svc.Create(context.Background(), model.CreateUserRequest{
		Name:  "Alex",
		Email: "alex@mail.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = svc.Create(context.Background(), model.CreateUserRequest{
		Name:  "Bob",
		Email: "alex@mail.com",
	})
	if err == nil {
		t.Fatal("expected conflict error")
	}

	var conflict *apperror.ConflictError
	if !errors.As(err, &conflict) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}
