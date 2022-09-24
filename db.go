package surrealdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const statusOK = "OK"

var (
	InvalidResponse = errors.New("invalid SurrealDB response")
	QueryError      = errors.New("error occurred processing the SurrealDB query")
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

// Unmarshal loads a SurrealDB response into a struct.
func Unmarshal(data interface{}, v interface{}) error {
	var ok bool

	assertedData, ok := data.([]interface{})
	if !ok {
		return InvalidResponse
	}
	sliceFlag := isSlice(v)

	var jsonBytes []byte
	var err error
	if !sliceFlag && len(assertedData) > 0 {
		jsonBytes, err = json.Marshal(assertedData[0])
	} else {
		jsonBytes, err = json.Marshal(assertedData)
	}
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, v)
	if err != nil {
		return err
	}

	return err
}

// UnmarshalRaw loads a raw SurrealQL response returned by Query into a struct. Queries that return with results will
// return ok = true, and queries that return with no results will return ok = false.
func UnmarshalRaw(rawData interface{}, v interface{}) (ok bool, err error) {
	var data []interface{}
	if data, ok = rawData.([]interface{}); !ok {
		return false, InvalidResponse
	}

	var responseObj map[string]interface{}
	if responseObj, ok = data[0].(map[string]interface{}); !ok {
		return false, InvalidResponse
	}

	var status string
	if status, ok = responseObj["status"].(string); !ok {
		return false, InvalidResponse
	}
	if status != statusOK {
		return false, QueryError
	}

	result := responseObj["result"]
	if len(result.([]interface{})) == 0 {
		return false, nil
	}
	err = Unmarshal(result, v)
	if err != nil {
		return false, err
	}

	return true, nil
}

// --------------------------------------------------
// Public methods
// --------------------------------------------------

// Close closes the underlying WebSocket connection.
func (database *DB) Close() error{
	return database.ws.Close()
}

// --------------------------------------------------

// Use is a method to select the namespace and table to use.
func (database *DB) Use(ns string, db string) (interface{}, error) {
	return database.send("use", ns, db)
}

func (database *DB) Info() (interface{}, error) {
	return database.send("info")
}

// Signup is a helper method for signing up a new user.
func (database *DB) Signup(vars interface{}) (interface{}, error) {
	return database.send("signup", vars)
}

// Signin is a helper method for signing in a user.
func (database *DB) Signin(vars interface{}) (interface{}, error) {
	return database.send("signin", vars)
}

func (database *DB) Invalidate() (interface{}, error) {
	return database.send("invalidate")
}

func (database *DB) Authenticate(token string) (interface{}, error) {
	return database.send("authenticate", token)
}

// --------------------------------------------------

func (database *DB) Live(table string) (interface{}, error) {
	return database.send("live", table)
}

func (database *DB) Kill(query string) (interface{}, error) {
	return database.send("kill", query)
}

func (database *DB) Let(key string, val interface{}) (interface{}, error) {
	return database.send("let", key, val)
}

// Query is a convenient method for sending a query to the database.
func (database *DB) Query(sql string, vars interface{}) (interface{}, error) {
	return database.send("query", sql, vars)
}

// Select a table or record from the database.
func (database *DB) Select(what string) (interface{}, error) {
	return database.send("select", what)
}

// Creates a table or record in the database like a POST request.
func (database *DB) Create(thing string, data interface{}) (interface{}, error) {
	return database.send("create", thing, data)
}

// Update a table or record in the database like a PUT request.
func (database *DB) Update(what string, data interface{}) (interface{}, error) {
	return database.send("update", what, data)
}

// Change a table or record in the database like a PATCH request.
func (database *DB) Change(what string, data interface{}) (interface{}, error) {
	return database.send("change", what, data)
}

// Modify applies a series of JSONPatches to a table or record.
func (database *DB) Modify(what string, data []Patch) (interface{}, error) {
	return database.send("modify", what, data)
}

// Delete a table or a row from the database like a DELETE request.
func (database *DB) Delete(what string) (interface{}, error) {
	return database.send("delete", what)
}

// --------------------------------------------------
// Private methods
// --------------------------------------------------

// send is a helper method for sending a query to the database.
func (database *DB) send(method string, params ...interface{}) (interface{}, error) {

	// generate an id for the action, this is used to distinguish its response
	id := xid(16)
	// chn: the channel where the server response will arrive, err: the channel where errors will come
	chn, err := database.ws.Once(id, method)
	// here we send the args through our websocket connection
	database.ws.Send(id, method, params)

	for {
		select {
		default:
		case e := <-err:
			return nil, e
		case r := <-chn:
			switch method {
			case "delete":
				return nil, nil
			//case "select":
			//	return self.resp(method, params, r)
			//case "create":
			//	return self.resp(method, params, r)
			//case "update":
			//	return self.resp(method, params, r)
			//case "change":
			//	return self.resp(method, params, r)
			//case "modify":
			//	return self.resp(method, params, r)
			default:
				return r, nil
			}
		}
	}

}

// resp is a helper method for parsing the response from a query.
func (database *DB) resp(_ string, _ []interface{}, result []byte) ([]byte, error) {

	fmt.Println("result:", string(result))

	return result, nil

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
