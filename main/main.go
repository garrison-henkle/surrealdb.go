package main

import (
	"context"
	"fmt"
	"github.com/surrealdb/surrealdb.go"
)

type testUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (t testUser) String() string {
	return fmt.Sprintf("testUser(id=%s, name=%s)", t.ID, t.Name)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := surrealdb.New(ctx, "ws://localhost:8000/rpc")
	if err != nil {
		panic(err)
	}
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(ctx, surrealdb.UserInfo{
		User:     "root",
		Password: "root",
	})

	err = db.Use(ctx, "test", "test")

	err = db.Delete(ctx, "testUser")

	var jim testUser
	err = db.Create(ctx, "testUser", map[string]interface{}{
		"name": "jim",
	}).Unmarshal(&jim)

	if err != nil {
		panic(err)
	}
	fmt.Println("jim 1:", jim)

	var jims []testUser
	err = db.Select(ctx, "testUser").Unmarshal(&jims)
	if err != nil {
		panic(err)
	}
	fmt.Println("jims:", jims)

	err = db.Select(ctx, "testUser").Unmarshal(&jim)
	if err != nil {
		panic(err)
	}
	fmt.Println("jim 2:", jim)

	err = db.Select(ctx, "testUser").Unmarshal(&jim)
	if err != nil && err != surrealdb.ErrNoResult {
		panic(err)
	}
	fmt.Println("is no result:", err == surrealdb.ErrNoResult)
	fmt.Println("no jims:", jim)

	jimmySend := testUser{
		Name: "Jimmy",
	}
	var jimmyReceive testUser
	err = db.Create(ctx, "testUser", &jimmySend).Unmarshal(&jimmyReceive)
	if err != nil {
		panic(err)
	}
	fmt.Println("jimmy sent:", jimmySend)
	fmt.Println("jimmy received:", jimmyReceive)

	jimmySend.ID = ""
	jimmySend.Name = "jimmy 2"
	err = db.Update(ctx, jimmyReceive.ID, &jimmySend).Unmarshal(&jimmyReceive)
	if err != nil {
		panic(err)
	}
	fmt.Println("jimmy 2 sent:", jimmySend)
	fmt.Println("jimmy 2 received:", jimmyReceive)

	var users []testUser
	err = db.Query(ctx, "select * from testUser", nil).Unmarshal(&users)
	if err != nil {
		panic(err)
	}
	fmt.Println("users:", users)

	var user testUser
	err = db.Query(ctx, "select * from testUser where id = "+jimmyReceive.ID, nil).Unmarshal(&user)
	if err != nil {
		panic(err)
	}
	fmt.Println("user:", user)

	var users2 []testUser
	err = db.Query(ctx, "select * from testUser where id = "+jimmyReceive.ID, nil).Unmarshal(&users2)
	if err != nil {
		panic(err)
	}
	fmt.Println("user in slice:", users2)

	var user3 testUser
	err = db.Query(ctx, "select * from testUser where name = 'jimmy'", nil).Unmarshal(&user3)
	if err != nil && err != surrealdb.ErrNoResult {
		panic(err)
	}
	fmt.Println("user 3:", user3)

	var user2 testUser
	err = db.Query(ctx, "selec t* from testUsr where name = 'jim'", nil).Unmarshal(&user2)
	if err == nil {
		panic("query should fail")
	}
	fmt.Println("user 2 query successfully failed")

}
