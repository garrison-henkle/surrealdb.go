package surrealdb_test

import (
	"context"
	"fmt"
	"github.com/garrison-henkle/surrealdb.go"
	"log"
	"testing"
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

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")

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

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(&ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(&ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	var user testUser
	err = db.Create(&ctx, "users", testUser{
		Username: "johnny",
		Password: "123",
	}).Unmarshal(&user)

	// Delete the users...
	err = db.Delete(&ctx, "users")

	if err != nil {
		panic(err)
	}

	// Output:
}

func TestDB_Create(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")

	if err != nil {
		panic(err)
	}

	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(&ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(&ctx, "test", "test")

	//todo, there was a check to see if signin was null here before the changes.
	// Maybe need a check for 'len(SurrealWSResult.Result) == 0' in the methods that just return error
	if err != nil {
		panic(err)
	}

	var john testUser
	err = db.Create(&ctx, "users", map[string]interface{}{
		"username": "john",
		"password": "123",
	}).Unmarshal(&john)

	if err != nil {
		panic(err)
	}

	var johnny testUser
	err = db.Create(&ctx, "users", testUser{
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

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")

	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(&ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(&ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	err = db.Create(&ctx, "users", testUser{
		Username: "johnnyjohn",
		Password: "123",
	}).Error

	// unmarshal the data into a user slice
	var users []testUser
	response := db.Select(&ctx, "users")
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

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(&ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(&ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	// unmarshal the data into a user struct
	var user testUser
	err = db.Create(&ctx, "users", testUser{
		Username: "johnny",
		Password: "123",
	}).Unmarshal(&user)

	if err != nil {
		panic(err)
	}

	user.Password = "456"

	// Update the user and unmarshal the data into a user struct
	var updatedUser []testUser
	err = db.Update(&ctx, "users", &user).Unmarshal(&updatedUser)

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

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(&ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(&ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	err = db.Delete(&ctx, "users")
	if err != nil {
		panic(err)
	}

	username := "johnny"
	password := "123"

	//create test user with raw SurrealQL and unmarshal

	var user testUser
	err = db.Query(&ctx, "create users:johnny set Username = $user, Password = $pass", map[string]interface{}{
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
	err = db.Query(&ctx, "select * from users where id = $id", map[string]interface{}{
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

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(&ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})
	err = db.Use(&ctx, "test", "test")

	_ = db.Delete(&ctx, "users")

	err = db.Create(&ctx, "users:999", map[string]interface{}{
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
	err = db.Modify(&ctx, "users:999", patches).Error
	if err != nil {
		panic(err)
	}

	var user map[string]interface{}
	err = db.Select(&ctx, "users:999").Unmarshal(&user)
	if err != nil {
		panic(err)
	}

	// // TODO: this needs to simplified for the end user somehow
	fmt.Println(user["age"])
	//
	// Output: 44
}

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// todo - temp to make sure the new stuff works
func TestUnmarshals(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(&ctx, "ws://localhost:8000/rpc")

	if err != nil {
		panic(err)
	}

	//wrap in closure to avoid unhandled error warnings
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(&ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	if err != nil {
		panic(err)
	}

	err = db.Use(&ctx, "test", "test")

	if err != nil {
		panic(err)
	}

	err = db.Delete(&ctx, "user")

	if err != nil {
		panic(err)
	}

	var user0 user
	err = db.Create(&ctx, "user:2", &user{
		Name: "garrison",
	}).Unmarshal(&user0)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 0 error:", err)
		panic(err)
	} else {
		fmt.Println("query 0:", user0)
	}

	var user1 user
	err = db.Create(&ctx, "user:1", &user{
		Name: "garrison",
	}).Unmarshal(&user1)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 1 error:", err)
		panic(err)
	} else {
		fmt.Println("query 1:", user1)
	}

	var users1 []user
	err = db.Select(&ctx, "user").Unmarshal(&users1)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 2 error:", err)
		panic(err)
	} else {
		fmt.Println("query 2:", users1)
	}

	var user2 user
	err = db.Update(&ctx, "user:1", &user{
		Name: "garrison 2",
	}).Unmarshal(&user2)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 3 error:", err)
		panic(err)
	} else {
		fmt.Println("query 3:", user2)
	}

	var users2a []user
	err = db.Query(&ctx, "select * from user", nil).Unmarshal(&users2a)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 4a error:", err)
		panic(err)
	} else {
		fmt.Println("query 4a:", users2a)
	}

	users2b, err := surrealdb.QueryAs[[]user](&ctx, db, "select * from user", nil)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 4b error:", err)
		panic(err)
	} else {
		fmt.Println("query 4b:", users2b)
	}

	var users3 []user
	err = db.Query(&ctx, "select * from usr", nil).Unmarshal(&users3)
	if err != nil {
		fmt.Println("query 5 error:", err)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 5:", users3)
		panic("err")
	}

	var users4 []user
	err = db.Query(&ctx, "selec t* from user", nil).Unmarshal(&users4)
	if err != nil {
		fmt.Println("query 6 error:", err)
	} else {
		fmt.Println("query 6:", users4)
		panic("err")
	}

	var users5 []user
	var users6 []user
	errs := db.Query(&ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&users5, &users6)
	if errs != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 7 error:", errs)
		panic(errs)
	} else {
		fmt.Println("query 7:\n - ", users5, "\n - ", users6)
	}

	var users7 []user
	var user3 user
	errs = db.Query(&ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&users7, &user3)
	if errs != nil {
		fmt.Println("query 8 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		//note: this is expected if the select only returns 1 record
		fmt.Println("query 8:\n - ", users7, "\n - ", user3)
		panic("errs")
	}

	var users8 []user
	var users9 []user
	errs = db.Query(&ctx, "select * from user; selec t* from user;", nil).UnmarshalMultiQuery(&users8, &users9)
	if errs != nil {
		fmt.Println("query 9 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 9:\n - ", users8, "\n - ", users9)
		panic("errs")
	}

	var user4 user
	errs = db.Query(&ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&user4)
	if errs != nil {
		fmt.Println("query 10 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 10:\n - ", users4)
		panic("errs")
	}

	var user5 user
	var user6 user
	var user7 user
	errs = db.Query(&ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&user5, &user6, &user7)
	if errs != nil {
		fmt.Println("query 11 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 11:\n - ", user5, "\n - ", user6, "\n - ", user7)
		panic("errs")
	}

	var users10 []user
	var users11 []user
	var users12 []user
	errs = db.Query(&ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&users10, &users11, &users12)
	if errs != nil {
		fmt.Println("query 12 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 12:\n - ", users10, "\n - ", users11, "\n - ", users12)
		panic("errs")
	}

	var users13 user
	err = db.Query(&ctx, "select * from user; select * from user;", nil).Unmarshal(&users13)
	if err != nil {
		fmt.Println("query 13 error:", err)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		//note: this is expected if the select only returns 1 record
		fmt.Println("query 13:", users13)
		panic("errs")
	}

	var users14 []user
	var users15 []user
	errs = db.Query(&ctx, "select * from user", nil).UnmarshalMultiQuery(&users14, &users15)
	if errs != nil {
		fmt.Println("query 14 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 14:\n - ", users14, "\n - ", users15)
		panic("errs")
	}

	var users16 []user
	var users17 []user
	errs = db.Query(&ctx, "begin transaction; select * from user; delete from user; select * from user; commit transaction;", nil).UnmarshalMultiQuery(&users16, nil, &users17)
	if errs != nil {
		fmt.Println("query 15 error:", errs)

	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 15:\n - ", users16, "\n - ", users17)
		panic("errs")
	}

}
