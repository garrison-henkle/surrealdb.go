package surrealdb

import (
	"errors"
	"fmt"
)

var (
	ErrNoResult          = errors.New("query returned no results")
	ErrTooFewContainers  = errors.New("a result has no container to unmarshal into")
	ErrTooManyContainers = errors.New("there were not enough results to fill the provided containers")
)

type ErrFailedRawQuery struct {
	Cause error
}

func (e ErrFailedRawQuery) Error() string {
	return generateErrorMsg("the raw query returned with an error", e.Cause)
}

type ErrInvalidSurrealResponse struct {
	Cause error
}

func (e ErrInvalidSurrealResponse) Error() string {
	return generateErrorMsg("the SurrealDB response is invalid and cannot be parsed", e.Cause)
}

type ErrFailedUnmarshal struct {
	Cause error
}

func (e ErrFailedUnmarshal) Error() string {
	return generateErrorMsg("could not unmarshal into provided container", e.Cause)
}

func generateErrorMsg(base string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s. Cause: %s", base, err.Error())
	}
	return base
}
