package surrealdb

import (
	"errors"
	"fmt"
)

var (
	ErrNoResult      = errors.New("query returned no results")
	ErrBlankResponse = errors.New("response was blank")
)

type ErrInvalidSurrealResponse struct {
	Cause error
}

func (e ErrInvalidSurrealResponse) Error() string {
	return generateErrorMsg("the SurrealDB response is invalid and cannot be parsed", e.Cause)
}

type ErrUnableToUnmarshal struct {
	Cause error
}

func (e ErrUnableToUnmarshal) Error() string {
	return generateErrorMsg("unable to marshal into provided container", e.Cause)
}

func generateErrorMsg(base string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s. Cause: %s", base, err.Error())
	}
	return base
}
