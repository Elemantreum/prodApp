package apperrors

import "fmt"

type UnauthorizedError struct {
}

func (e UnauthorizedError) Error() string {
	return "unauthorized"
}

type ForbiddenError struct {
	Reason string
}

func (f ForbiddenError) Error() string {
	if f.Reason == "" {
		return "forbidden"
	}
	return fmt.Sprintf("forbidden: %s", f.Reason)
}

type ValidationError struct {
	Field string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("invalid field: %s", e.Field)
}

type NotFoundError struct {
}

func (e NotFoundError) Error() string {
	return "not found"
}

type ConflictError struct {
	Resource string
	Reason   string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("%s: %s", e.Resource, e.Reason)
}
