package main

import (
	"context"
	"fmt"
	"github.com/surrealdb/surrealdb.go"
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
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

	err = db.Delete(ctx, "user")

	if err != nil {
		panic(err)
	}

	var user0 user
	err = db.Create(ctx, "user:2", &user{
		Name: "garrison",
	}).Unmarshal(&user0)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 0 error:", err)
	} else {
		fmt.Println("query 0:", user0)
	}

	var user1 user
	err = db.Create(ctx, "user:1", &user{
		Name: "garrison",
	}).Unmarshal(&user1)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 1 error:", err)
	} else {
		fmt.Println("query 1:", user1)
	}

	var users1 []user
	err = db.Select(ctx, "user").Unmarshal(&users1)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 2 error:", err)
	} else {
		fmt.Println("query 2:", users1)
	}

	var user2 user
	err = db.Update(ctx, "user:1", &user{
		Name: "garrison 2",
	}).Unmarshal(&user2)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 3 error:", err)
	} else {
		fmt.Println("query 3:", user2)
	}

	var users2 []user
	err = db.Query(ctx, "select * from user", nil).Unmarshal(&users2)
	if err != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 4 error:", err)
	} else {
		fmt.Println("query 4:", users2)
	}

	var users3 []user
	err = db.Query(ctx, "select * from usr", nil).Unmarshal(&users3)
	if err != nil {
		fmt.Println("query 5 error:", err)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 5:", users3)
	}

	var users4 []user
	err = db.Query(ctx, "selec t* from user", nil).Unmarshal(&users4)
	if err != nil {
		fmt.Println("query 6 error:", err)
	} else {
		fmt.Println("query 6:", users4)
	}

	var users5 []user
	var users6 []user
	errs := db.Query(ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&users5, &users6)
	if errs != nil {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 7 error:", errs)
	} else {
		fmt.Println("query 7:\n - ", users5, "\n - ", users6)
	}

	var users7 []user
	var user3 user
	errs = db.Query(ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&users7, &user3)
	if errs != nil {
		fmt.Println("query 8 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		//note: this is expected if the select only returns 1 record
		fmt.Println("query 8:\n - ", users7, "\n - ", user3)
	}

	var users8 []user
	var users9 []user
	errs = db.Query(ctx, "select * from user; selec t* from user;", nil).UnmarshalMultiQuery(&users8, &users9)
	if errs != nil {
		fmt.Println("query 9 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 9:\n - ", users8, "\n - ", users9)
	}

	var user4 user
	errs = db.Query(ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&user4)
	if errs != nil {
		fmt.Println("query 10 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 10:\n - ", users4)
	}

	var user5 user
	var user6 user
	var user7 user
	errs = db.Query(ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&user5, &user6, &user7)
	if errs != nil {
		fmt.Println("query 11 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 11:\n - ", user5, "\n - ", user6, "\n - ", user7)
	}

	var users10 []user
	var users11 []user
	var users12 []user
	errs = db.Query(ctx, "select * from user; select * from user;", nil).UnmarshalMultiQuery(&users10, &users11, &users12)
	if errs != nil {
		fmt.Println("query 12 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 12:\n - ", users10, "\n - ", users11, "\n - ", users12)
	}

	var users13 user
	err = db.Query(ctx, "select * from user; select * from user;", nil).Unmarshal(&users13)
	if err != nil {
		fmt.Println("query 13 error:", err)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		//note: this is expected if the select only returns 1 record
		fmt.Println("query 13:", users13)
	}

	var users14 []user
	var users15 []user
	errs = db.Query(ctx, "select * from user", nil).UnmarshalMultiQuery(&users14, &users15)
	if errs != nil {
		fmt.Println("query 14 error:", errs)
	} else {
		_ = fmt.Errorf("UNEXPECTED RESULT: ")
		fmt.Println("query 14:\n - ", users14, "\n - ", users15)
	}
}
