package surrealdb

import "fmt"

// Patch represents a patch object set to MODIFY a record
type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type UserInfo struct {
	User     string `json:"user"`
	Password string `json:"pass"`
}

type MultiQueryError struct {
	QueryNumber int
	Error       error
}

type SurrealWSResult struct {
	Result        []byte
	SingleRequest bool
	Error         error
}

func (r SurrealWSResult) String() string {
	if r.Error != nil {
		return fmt.Sprintf("SurrealWSResult(error=%s)", r.Error.Error())
	}
	return fmt.Sprintf("SurrealWSResult(data=%s, single=%v)", r.Result, r.SingleRequest)
}

type SurrealWSRawResult struct {
	Result []byte
	Error  error
}

func (r SurrealWSRawResult) String() string {
	if r.Error != nil {
		return fmt.Sprintf("SurrealWSRawResult(error=%s)", r.Error.Error())
	}
	return fmt.Sprintf("SurrealWSRawResult(data=%s)", r.Result)
}
