package surrealdb

import (
	"fmt"
)

type ErrNoPermission struct {
	what string
}

func (e ErrNoPermission) Error() string {
	return fmt.Sprint("Unable to access record:", e.what)
}

type ErrInvalidSurrealResponse struct {
	Cause error
}

func (e ErrInvalidSurrealResponse) Error() string {
	msg := "The SurrealDB response is invalid and cannot be parsed"
	if e.Cause != nil {
		return fmt.Sprintf("%s. Cause: %s", msg, e.Cause.Error())
	}
	return msg
}
