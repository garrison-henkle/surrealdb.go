package surrealdb

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DB is a client for the SurrealDB database that holds are websocket connection.
type DB struct {
	ws *WS
}

// New Creates a new DB instance given a WebSocket URL.
func New(url string) (*DB, error) {
	ws, err := NewWebsocket(url)
	if err != nil {
		return nil, err
	}
	return &DB{ws}, nil
}

// --------------------------------------------------

type SurrealWSResult struct {
	Result []byte
	Single bool
	Error  error
}

type SurrealWSRawResult struct {
	Result []byte
	Error  error
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
	if !r.Single && isSlice(v) {
		jsonBytes = r.Result
	} else {
		jsonBytes = r.Result[1:(resultLength - 1)]
	}
	err := json.Unmarshal(jsonBytes, v)
	if err != nil {
		return ErrInvalidSurrealResponse{Cause: err}
	}
	return nil
}

func (r SurrealWSRawResult) Unmarshal(v interface{}) error {
	if r.Error != nil {
		return r.Error
	}

	responseLength := len(r.Result)
	if responseLength <= 2 {
		return ErrInvalidSurrealResponse{}
	}
	var rawJson map[string]json.RawMessage
	err := json.Unmarshal(r.Result[1:(responseLength-1)], &rawJson)
	if err != nil {
		return ErrInvalidSurrealResponse{Cause: err}
	}

	rpcErrBytes, errOccurred := rawJson["error"]
	if errOccurred {
		var rpcErr RPCError
		err = json.Unmarshal(rpcErrBytes, &rpcErr)
		if err != nil {
			return ErrInvalidSurrealResponse{Cause: err}
		}
		return &rpcErr
	}

	var result []byte
	result, ok := rawJson["result"]
	if !ok {
		return ErrInvalidSurrealResponse{}
	}
	resultLength := len(result)

	//check for empty result
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
		return ErrUnableToUnmarshal{Cause: err}
	}
	return nil
}

// --------------------------------------------------
// Public methods
// --------------------------------------------------

// Close closes the underlying WebSocket connection.
func (database *DB) Close() error {
	return database.ws.Close()
}

// --------------------------------------------------

// Use is a method to select the namespace and table to use.
func (database *DB) Use(ns string, db string) (err error) {
	return database.send("use", ns, db).Error
}

func (database *DB) Info() (err error) {
	return database.send("info").Error
}

// Signup is a helper method for signing up a new user.
func (database *DB) Signup(vars interface{}) (err error) {
	return database.send("signup", vars).Error
}

// Signin is a helper method for signing in a user.
func (database *DB) Signin(vars interface{}) (err error) {
	return database.send("signin", vars).Error
}

func (database *DB) Invalidate() (err error) {
	return database.send("invalidate").Error
}

func (database *DB) Authenticate(token string) (err error) {
	return database.send("authenticate", token).Error
}

// --------------------------------------------------

func (database *DB) Live(table string) (err error) {
	return database.send("live", table).Error
}

func (database *DB) Kill(query string) (err error) {
	return database.send("kill", query).Error
}

func (database *DB) Let(key string, val interface{}) (err error) {
	return database.send("let", key, val).Error
}

// Query is a convenient method for sending a query to the database.
func (database *DB) Query(sql string, vars interface{}) *SurrealWSRawResult {
	return database.sendRaw("query", sql, vars)
}

// Select a table or record from the database.
func (database *DB) Select(what string) *SurrealWSResult {
	return database.send("select", what)
}

// Create a table or record in the database like a POST request.
func (database *DB) Create(thing string, data interface{}) *SurrealWSResult {
	return database.send("create", thing, data)
}

// Update a table or record in the database like a PUT request.
func (database *DB) Update(what string, data interface{}) *SurrealWSResult {
	return database.send("update", what, data)
}

// Change a table or record in the database like a PATCH request.
func (database *DB) Change(what string, data interface{}) *SurrealWSResult {
	return database.send("change", what, data)
}

// Modify applies a series of JSONPatches to a table or record.
func (database *DB) Modify(what string, data []Patch) *SurrealWSResult {
	return database.send("modify", what, data)
}

// Delete a table or a row from the database like a DELETE request.
func (database *DB) Delete(what string) error {
	return database.send("delete", what).Error
}

// --------------------------------------------------
// Private methods
// --------------------------------------------------

// send is a helper method for sending a query to the database.
func (database *DB) send(method string, params ...interface{}) *SurrealWSResult {
	response, err := sendMessage(database.ws, method, params)
	for {
		select {
		default:
		case e := <-err:
			return &SurrealWSResult{Error: e}
		case result := <-response:
			arg, ok := params[0].(string)
			singleRecordRequested := ok && strings.Contains(arg, ":")
			return &SurrealWSResult{
				Result: result,
				Single: singleRecordRequested,
			}
		}
	}
}

func (database *DB) sendRaw(method string, params ...interface{}) *SurrealWSRawResult {
	response, err := sendMessage(database.ws, method, params)
	for {
		select {
		default:
		case e := <-err:
			return &SurrealWSRawResult{Error: e}
		case result := <-response:
			return &SurrealWSRawResult{
				Result: result,
			}
		}
	}
}

func sendMessage(ws *WS, method string, params []interface{}) (responseChannel <-chan []byte, errorChannel <-chan error) {
	// generate an id for the action, this is used to distinguish its response
	id := xid(16)
	// chn: the channel where the server response will arrive, err: the channel where errors will come
	responseChannel, errorChannel = ws.Once(id, method)
	// here we send the args through our websocket connection
	ws.Send(id, method, params)
	return responseChannel, errorChannel
}

func isSlice(possibleSlice interface{}) bool {
	slice := false

	switch v := possibleSlice.(type) {
	default:
		res := fmt.Sprintf("%s", v)
		if res == "[]" || res == "&[]" || res == "*[]" {
			slice = true
		}
	}

	return slice
}
