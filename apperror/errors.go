package apperror

import "fmt"

type NotFoundError struct {
	Entity string
	ID     int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %d not found", e.Entity, e.ID)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation: %s — %s", e.Field, e.Message)
}

type ConflictError struct {
	Entity string
	Field  string
	Value  string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s with %s=%q already exists", e.Entity, e.Field, e.Value)
}
