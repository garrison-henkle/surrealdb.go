package surrealdb_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/surrealdb/surrealdb.go"
)

// a simple user struct for testing
type testUser struct {
	Username string
	Password string
	ID       string
}

// an example test for creating a new entry in surrealdb
func TestNew(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")

	if err != nil {
		panic(err)
	}

	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	// Output:
}

func TestDB_Delete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	var user testUser
	err = db.Create(ctx, "users", testUser{
		Username: "johnny",
		Password: "123",
	}).Unmarshal(&user)

	// Delete the users...
	err = db.Delete(ctx, "users")

	if err != nil {
		panic(err)
	}

	// Output:
}

func TestDB_Create(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")

	if err != nil {
		panic(err)
	}

	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(ctx, "test", "test")

	//todo, there was a check to see if signin was null here before the changes.
	// Maybe need a check for 'len(SurrealWSResult.Result) == 0' in the methods that just return error
	if err != nil {
		panic(err)
	}

	var john testUser
	err = db.Create(ctx, "users", map[string]interface{}{
		"username": "john",
		"password": "123",
	}).Unmarshal(&john)

	if err != nil {
		panic(err)
	}

	var johnny testUser
	err = db.Create(ctx, "users", testUser{
		Username: "johnny",
		Password: "123",
	}).Unmarshal(&johnny)

	if err != nil {
		panic(err)
	}

	fmt.Println(johnny.Username)

	// Output: johnny
}

func TestDB_Select(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")

	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	err = db.Create(ctx, "users", testUser{
		Username: "johnnyjohn",
		Password: "123",
	}).Error

	// unmarshal the data into a user slice
	var users []testUser
	response := db.Select(ctx, "users")
	log.Print(response.Result)

	err = response.Unmarshal(&users)
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		if user.Username == "johnnyjohn" {
			fmt.Println(user.Username)
			break
		}
	}
	// Output: johnnyjohn
}

func TestDB_Update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	// unmarshal the data into a user struct
	var user testUser
	err = db.Create(ctx, "users", testUser{
		Username: "johnny",
		Password: "123",
	}).Unmarshal(&user)

	if err != nil {
		panic(err)
	}

	user.Password = "456"

	// Update the user and unmarshal the data into a user struct
	var updatedUser []testUser
	err = db.Update(ctx, "users", &user).Unmarshal(&updatedUser)

	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	// TODO: check if this updates only the user with the same ID or all users
	fmt.Println(updatedUser[0].Password)

	// Output: 456
}

func TestUnmarshalRaw(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	err = db.Delete(ctx, "users")
	if err != nil {
		panic(err)
	}

	username := "johnny"
	password := "123"

	//create test user with raw SurrealQL and unmarshal

	var user testUser
	err = db.Query(ctx, "create users:johnny set Username = $user, Password = $pass", map[string]interface{}{
		"user": username,
		"pass": password,
	}).Unmarshal(&user)
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}
	if user.Username != username || user.Password != password {
		panic("response does not match the request")
	}

	//send query with empty result and unmarshal

	var emptyUser testUser
	err = db.Query(ctx, "select * from users where id = $id", map[string]interface{}{
		"id": "users:jim",
	}).Unmarshal(&emptyUser)
	if err == nil || err != surrealdb.ErrNoResult {
		panic(err)
	}

	// Output:
}

func TestDB_Modify(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})
	err = db.Use(ctx, "test", "test")

	_ = db.Delete(ctx, "users")

	err = db.Create(ctx, "users:999", map[string]interface{}{
		"username": "john999",
		"password": "123",
	}).Error
	if err != nil {
		panic(err)
	}

	patches := []surrealdb.Patch{
		{Op: "add", Path: "nickname", Value: "johnny"},
		{Op: "add", Path: "age", Value: 44},
	}

	// Update the user
	err = db.Modify(ctx, "users:999", patches).Error
	if err != nil {
		panic(err)
	}

	var user map[string]interface{}
	err = db.Select(ctx, "users:999").Unmarshal(&user)
	if err != nil {
		panic(err)
	}

	// // TODO: this needs to simplified for the end user somehow
	fmt.Println(user["age"])
	//
	// Output: 44
}
