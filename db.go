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

// --------------------------------------------------
// Public methods
// --------------------------------------------------

// Close closes the underlying WebSocket connection.
func (database *DB) Close() error{
	return database.ws.Close()
}

// --------------------------------------------------

// Use is a method to select the namespace and table to use.
func (database *DB) Use(ns string, db string) (err error) {
	_, err = database.send("use", nil, ns, db)
	return err
}

func (database *DB) Info() (err error) {
	_, err = database.send("info", nil)
	return err
}

// Signup is a helper method for signing up a new user.
func (database *DB) Signup(vars interface{}) (err error) {
	_, err = database.send("signup", nil, vars)
	return err
}

// Signin is a helper method for signing in a user.
func (database *DB) Signin(vars interface{}) (err error) {
	_, err = database.send("signin", nil, vars)
	return err
}

func (database *DB) Invalidate() (err error) {
	_, err = database.send("invalidate", nil)
	return err
}

func (database *DB) Authenticate(token string) (err error) {
	_, err = database.send("authenticate", nil, token)
	return err
}

// --------------------------------------------------

func (database *DB) Live(table string) (err error) {
	_, err = database.send("live", nil, table)
	return err
}

func (database *DB) Kill(query string) (err error) {
	_, err = database.send("kill", nil, query)
	return err
}

func (database *DB) Let(key string, val interface{}) (err error) {
	_, err = database.send("let", nil, key, val)
	return err
}

// Query is a convenient method for sending a query to the database.
func (database *DB) Query(sql string, response interface{}, vars interface{}) (bool, error) {
	return database.send("query", response, sql, vars)
}

// Select a table or record from the database.
func (database *DB) Select(what string, response interface{}) (bool, error) {
	return database.send("select", response, what)
}

// Create a table or record in the database like a POST request.
func (database *DB) Create(thing string, data interface{}, response interface{}) (bool, error){
	return database.send("create", response, thing, data)
}

// Update a table or record in the database like a PUT request.
func (database *DB) Update(what string, data interface{}, response interface{}) (bool, error) {
	return database.send("update", response, what, data)
}

// Change a table or record in the database like a PATCH request.
func (database *DB) Change(what string, data interface{}, response interface{}) (bool, error) {
	return database.send("change", response, what, data)
}

// Modify applies a series of JSONPatches to a table or record.
func (database *DB) Modify(what string, data []Patch, response interface{}) (bool, error) {
	return database.send("modify", response, what, data)
}

// Delete a table or a row from the database like a DELETE request.
func (database *DB) Delete(what string, response interface{}) (bool, error) {
	return database.send("delete", response, what)
}

// --------------------------------------------------
// Private methods
// --------------------------------------------------

// send is a helper method for sending a query to the database.
func (database *DB) send(method string, container interface{}, params ...interface{}) (bool, error) {

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
			return false, e
		case r := <-chn:
			switch method {
			case "delete":
				return true, nil
			case "query":
				return database.unmarshalRaw(params, r, container)
			case "select":
				return database.unmarshal(params, r, container)
			case "create":
				return database.unmarshal(params, r, container)
			case "update":
				return database.unmarshal(params, r, container)
			case "change":
				return database.unmarshal(params, r, container)
			case "modify":
				return database.unmarshal(params, r, container)
			default:
				return true, nil
			}
		}
	}
}

const (
	leftBracket = 91
	rightBracket = 93
)
//[{"result":[],
//[{"result":[{"id":"testUser1:mhrzesa4616vv14xdq60","name":"jimbo"}],"status":"OK","time":"75.5µs"}]
//0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890

func getLastResultIndex(result []byte) (bool, int) {
	resultLength := len(result)
	var endIndex int
	for endIndex = resultLength - 2; endIndex >= 0; endIndex-- {
		if result[endIndex] == rightBracket {
			break
		}
	}
	if endIndex == 0{
		return false, 0
	}
	return true, endIndex
}

func (database *DB) unmarshalRaw(params []interface{}, result []byte, container interface{}) (bool, error){
	//first 11 characters are '[{"result":', so start on character 12 (index 11)
	startIndex := 11
	ok, endIndex := getLastResultIndex(result)

	//check for empty result
	if !ok || endIndex - startIndex - 1 <= 0 {
		return false, nil
	}

	var jsonBytes []byte
	var err error
	if isSlice(container){
		jsonBytes = result[startIndex:(endIndex + 1)]
	} else{
		ok, err = parseIfSingleObject(params, result, startIndex, endIndex + 1, container)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
		jsonBytes = result[(startIndex + 1):endIndex]
	}
	err = json.Unmarshal(jsonBytes, container)
	if err != nil{
		return false, err
	}
	return true, nil
}

func parseIfSingleObject(params []interface{}, result []byte, start int, end int, container interface{}) (bool, error){
	//check for empty result
	if end - start - 1 <= 0 {
		return false, nil
	}

	//if the query was not for a specific record, return
	arg, ok := params[0].(string)
	if !ok || !strings.Contains(arg, ":"){
		return false, nil
	}

	//strip the square brackets and marshal
	jsonBytes := result[(start + 1):(end - 1)]
	err := json.Unmarshal(jsonBytes, container)
	if err != nil{
		return false, err
	}
	return true, nil
}

// resp is a helper method for parsing the response from a query.
func (database *DB) unmarshal(params []interface{}, result []byte, container interface{}) (bool, error) {
	resultLength := len(result)

	//check for empty result
	if resultLength - 2 <= 0 {
		return false, nil
	}

	ok, err := parseIfSingleObject(params, result, 0, resultLength, container)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	var jsonBytes []byte
	if isSlice(container){
		jsonBytes = result
	} else{
		jsonBytes = result[1:(resultLength - 1)]
	}
	err = json.Unmarshal(jsonBytes, container)
	if err != nil{
		return false, err
	}
	return true, nil

	//arg, ok := params[0].(string)
	//
	//if ok && strings.Contains(arg, ":") {
	//	resultLength := len(result)
	//	jsonBytes := result[1:(resultLength - 1)]
	//	return json.Unmarshal(jsonBytes, container)
	//}
	//
	//return json.Unmarshal(result, container)

	//marshal as array


	//resultLength := len(result)
	//if resultLength > 0{
	//	first := result[0]
	//
	//	//if array (91 is '[')
	//	var i int
	//	if first == leftBracket{
	//		for i = resultLength - 1; i >= 0; i-- {
	//			if result[i] == comma && resultLength - i >= 7 && string(result[(i + 1):(i + 7)]) == time {
	//				break
	//			} else{
	//
	//			}
	//		}
	//	}
	//}

	//fmt.Println("resp:", string(result))
	//var resultMap map[string]json.RawMessage
	//err := json.Unmarshal(result, &resultMap)
	//if err != nil{
	//	return err
	//}
	//
	//fmt.Println("results:", &resultMap)


	//_, _ := params[0].(string)

	//if !ok {
	//	return result
	//}


	//if strings.Contains(arg, ":") {
	//
	//	arr, ok := result.([]interface{})
	//
	//	if !ok {
	//		return nil, PermissionError{what: arg}
	//	}
	//
	//	if len(arr) < 1 {
	//		return nil, PermissionError{what: arg}
	//	}
	//
	//	return arr[0], nil
	//
	//}

	//return nil
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
