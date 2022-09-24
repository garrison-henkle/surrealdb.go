package main

import (
	"fmt"
	"github.com/surrealdb/surrealdb.go"
)

func main(){
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil{
		panic(err)
	}
	defer db.Close()

	_, err = db.Signin(map[string]interface{}{
		"user": "root",
		"pass": "root",
	})
	_, err = db.Use("test", "test")

	_, err = db.Delete("testUser1")

	_, err = db.Create("testUser1", map[string]interface{}{
		"name": "jimbo",
	})

	data, err := db.Query("select * from testUser1 where name = 'jimbo'", nil)
	if err != nil{
		panic(err)
	}
	fmt.Println("final data 1: ", string(data.([]byte)))

	data, err = db.Select("testUser1")
	fmt.Println("final data 2: ", string(data.([]byte)))

	data, err = db.Query("selec t* from testUsr where name = 'jimbo'", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("final data 3: ", data)
}
