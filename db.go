package surrealdb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"strings"
)

// DB is a client for the SurrealDB database that holds are websocket connection.
type DB struct {
	ws *WS
}

// New Creates a new DB instance given a WebSocket URL.
func New(ctx context.Context, url string) (*DB, error) {
	ws, err := NewWebsocket(ctx, url)
	if err != nil {
		return nil, err
	}
	return &DB{ws}, nil
}

// --------------------------------------------------

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

func (r SurrealWSResult) Unmarshal(v interface{}) error {
	if r.Error != nil {
		return r.Error
	}

	resultLength := len(r.Result)

	//check for empty result
	if (resultLength - 2) <= 0 {
		return ErrNoResult
	}

	var jsonBytes []byte
	if !r.SingleRequest && isSlice(v) {
		jsonBytes = r.Result
	} else {
		jsonBytes = r.Result[1:(resultLength - 1)]
	}
	err := json.Unmarshal(jsonBytes, v)
	if err != nil {
		return ErrFailedUnmarshal{Cause: err}
	}
	return nil
}

type MultiQueryError struct {
	QueryNumber int
	Error       error
}

// Unmarshal unmarshals a single response returned by a raw query
func (r SurrealWSRawResult) Unmarshal(v interface{}) error {
	if r.Error != nil {
		return r.Error
	}

	result, _, _, err := jsonparser.Get(r.Result, "result")
	if err != nil {
		return ErrInvalidSurrealResponse{Cause: err}
	}

	//check for empty result
	resultLength := len(result)
	if (resultLength - 2) <= 0 {
		return ErrNoResult
	}

	result, _, _, err = jsonparser.Get(result[1:(resultLength-1)], "result")
	if err != nil {
		return ErrInvalidSurrealResponse{Cause: err}
	}

	//check for empty result
	resultLength = len(result)
	if (resultLength - 2) <= 0 {
		return ErrNoResult
	}

	var jsonBytes []byte
	if isSlice(v) {
		jsonBytes = result
	} else {
		jsonBytes = result[1:(resultLength - 1)]
	}
	err = json.Unmarshal(jsonBytes, v)
	if err != nil {
		return ErrFailedUnmarshal{Cause: err}
	}
	return nil
}

// UnmarshalMultiQuery unmarshals the response returned by queries sent in bulk into the provided containers.
func (r SurrealWSRawResult) UnmarshalMultiQuery(v ...interface{}) []MultiQueryError {
	containerCount := len(v)
	errorSlice := make([]MultiQueryError, 0)
	if r.Error != nil {
		errorSlice = append(errorSlice, MultiQueryError{
			QueryNumber: -1,
			Error:       r.Error,
		})
		return errorSlice
	}

	var jsonBytes []byte
	var resultLength int
	var queryNumber int
	fatalErrorOccurred := false
	_, err := jsonparser.ArrayEach(r.Result, func(result []byte, _ jsonparser.ValueType, _ int, err error) {
		if fatalErrorOccurred {
			return
		}
		if queryNumber >= containerCount {
			errorSlice = append(errorSlice, MultiQueryError{
				QueryNumber: -1,
				Error:       ErrTooFewContainers,
			})
			fatalErrorOccurred = true
			return
		}

		//check for empty result
		resultLength = len(result)
		if (resultLength - 2) <= 0 {
			errorSlice = append(errorSlice, MultiQueryError{
				QueryNumber: queryNumber,
				Error:       ErrNoResult,
			})
			queryNumber++
			return
		}

		result, _, _, err = jsonparser.Get(result, "result")
		if err != nil {
			errorSlice = append(errorSlice, MultiQueryError{
				QueryNumber: queryNumber,
				Error:       ErrInvalidSurrealResponse{Cause: err},
			})
			return
		}

		//check for empty result
		resultLength = len(result)
		if (resultLength - 2) <= 0 {
			errorSlice = append(errorSlice, MultiQueryError{
				QueryNumber: queryNumber,
				Error:       ErrNoResult,
			})
			queryNumber++
			return
		}

		if isSlice(v[queryNumber]) {
			jsonBytes = result
		} else {
			jsonBytes = result[1:(resultLength - 1)]
		}

		err = json.Unmarshal(jsonBytes, v[queryNumber])
		if err != nil {
			errorSlice = append(errorSlice, MultiQueryError{
				QueryNumber: queryNumber,
				Error:       &ErrFailedUnmarshal{Cause: err},
			})
		}

		queryNumber++
	}, "result")

	if err != nil {
		errorSlice = append(errorSlice, MultiQueryError{
			QueryNumber: -1,
			Error:       ErrInvalidSurrealResponse{Cause: err},
		})
		return errorSlice
	}

	if queryNumber != containerCount {
		errorSlice = append(errorSlice, MultiQueryError{
			QueryNumber: -1,
			Error:       ErrTooManyContainers,
		})
		return errorSlice
	}

	if len(errorSlice) != 0 {
		return errorSlice
	}
	return nil
}

// --------------------------------------------------
// Public methods
// --------------------------------------------------

// Close closes the underlying WebSocket connection.
func (db *DB) Close() error {
	return db.ws.Close()
}

// --------------------------------------------------

// Use is a method to select the namespace and table to use.
func (db *DB) Use(ctx context.Context, ns string, dbname string) error {
	return db.send(ctx, "use", ns, dbname).Error
}

func (db *DB) Info(ctx context.Context) error {
	return db.send(ctx, "info").Error
}

// Signup is a helper method for signing up a new user.
func (db *DB) Signup(ctx context.Context, vars interface{}) error {
	return db.send(ctx, "signup", vars).Error
}

// Signin is a helper method for signing in a user.
func (db *DB) Signin(ctx context.Context, vars UserInfo) error {
	return db.send(ctx, "signin", vars).Error
}

func (db *DB) Invalidate(ctx context.Context) error {
	return db.send(ctx, "invalidate").Error
}

func (db *DB) Authenticate(ctx context.Context, token string) error {
	return db.send(ctx, "authenticate", token).Error
}

// --------------------------------------------------

func (db *DB) Live(ctx context.Context, table string) error {
	return db.send(ctx, "live", table).Error
}

func (db *DB) Kill(ctx context.Context, query string) error {
	return db.send(ctx, "kill", query).Error
}

func (db *DB) Let(ctx context.Context, key string, val interface{}) error {
	return db.send(ctx, "let", key, val).Error
}

// Query is a convenient method for sending a query to the database.
func (db *DB) Query(ctx context.Context, sql string, vars interface{}) *SurrealWSRawResult {
	return db.sendRaw(ctx, "query", sql, vars)
}

// Select a table or record from the database.
func (db *DB) Select(ctx context.Context, what string) *SurrealWSResult {
	return db.send(ctx, "select", what)
}

// Create a table or record in the database like a POST request.
func (db *DB) Create(ctx context.Context, thing string, data interface{}) *SurrealWSResult {
	return db.send(ctx, "create", thing, data)
}

// Update a table or record in the database like a PUT request.
func (db *DB) Update(ctx context.Context, what string, data interface{}) *SurrealWSResult {
	return db.send(ctx, "update", what, data)
}

// Change a table or record in the database like a PATCH request.
func (db *DB) Change(ctx context.Context, what string, data interface{}) *SurrealWSResult {
	return db.send(ctx, "change", what, data)
}

// Modify applies a series of JSONPatches to a table or record.
func (db *DB) Modify(ctx context.Context, what string, data []Patch) *SurrealWSResult {
	return db.send(ctx, "modify", what, data)
}

// Delete a table or a row from the database like a DELETE request.
func (db *DB) Delete(ctx context.Context, what string) error {
	return db.send(ctx, "delete", what).Error
}

// --------------------------------------------------
// Private methods
// --------------------------------------------------

// send is a helper method for sending a query to the database.
func (db *DB) send(ctx context.Context, method string, params ...interface{}) *SurrealWSResult {
	response := sendMessage(db.ws, method, params)
	for {
		select {
		case <-ctx.Done():
			return &SurrealWSResult{Error: ctx.Err()}
		case r := <-response:
			resultBytes, err := parseResponse(r.Value)
			result := SurrealWSResult{
				Result: resultBytes,
				Error:  err,
			}
			if err == nil {
				arg, ok := params[0].(string)
				singleRecordRequested := ok && strings.Contains(arg, ":")
				result.SingleRequest = singleRecordRequested
			}
			return &result
		}
	}
}

func (db *DB) sendRaw(ctx context.Context, method string, params ...interface{}) *SurrealWSRawResult {
	response := sendMessage(db.ws, method, params)
	for {
		select {
		default:
		case <-ctx.Done():
			return &SurrealWSRawResult{Error: ctx.Err()}
		case r := <-response:
			resultBytes, err := parseRawResponse(r.Value)
			return &SurrealWSRawResult{
				Result: resultBytes,
				Error:  err,
			}
		}
	}
}

func parseRawResponse(response []byte) ([]byte, error) {
	data, _, _, err := jsonparser.Get(response, "error")
	if err != nil {
		return response, nil
	}
	var rpcErr RPCError
	err = json.Unmarshal(data, &rpcErr)
	if err != nil {
		return nil, &ErrInvalidSurrealResponse{Cause: err}
	}
	return nil, &rpcErr
}

func parseResponse(response []byte) ([]byte, error) {
	//try result first because it is more likely than an error
	data, _, _, err := jsonparser.Get(response, "result")
	if err == nil {
		return data, nil
	}

	data, _, _, err = jsonparser.Get(response, "error")
	if err != nil {
		return nil, &ErrInvalidSurrealResponse{Cause: err}
	}
	var rpcErr RPCError
	err = json.Unmarshal(data, &rpcErr)
	if err != nil {
		return nil, &ErrInvalidSurrealResponse{Cause: err}
	}
	return nil, &rpcErr
}

func sendMessage(ws *WS, method string, params []interface{}) (responseChannel <-chan responseValue) {
	// generate an id for the action, this is used to distinguish its response
	id := xid()
	// chn: the channel where the server response will arrive, err: the channel where errors will come
	responseChannel = ws.Once(id, method)
	// here we send the args through our websocket connection
	ws.Send(id, method, params)
	return responseChannel
}

func isSlice(possibleSlice interface{}) bool {
	res := fmt.Sprintf("%s", possibleSlice)
	if res == "[]" || res == "&[]" || res == "*[]" {
		return true
	}

	return false
}
