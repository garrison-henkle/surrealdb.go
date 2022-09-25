package main

import (
	"fmt"
	"github.com/surrealdb/surrealdb.go"
)

type testUser struct {
	ID string `json:"id"`
	Name string `json:"name"`
}

func (t testUser) String() string{
	return fmt.Sprintf("testUser(id=%s, name=%s)", t.ID, t.Name)
}

func main(){
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil{
		panic(err)
	}
	defer func(db *surrealdb.DB) {
		_ = db.Close()
	}(db)

	err = db.Signin(map[string]interface{}{
		"user": "root",
		"pass": "root",
	})

	err = db.Use("test", "test")

	_, err = db.Delete("testUser", nil)

	var ok bool
	var jim testUser
	ok, err = db.Create("testUser", map[string]interface{}{
		"name": "jim",
	}, &jim)

	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("jimbo response was empty")
	} else{
		fmt.Println("jim 1:", jim)
	}

	var jims []testUser
	ok, err = db.Select("testUser", &jims)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("jim 1 response was empty")
	} else{
		fmt.Println("jims:", jims)
	}

	ok, err = db.Select("testUser", &jim)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("jims response was empty")
	} else{
		fmt.Println("jim 2:", jim)
	}

	ok, err = db.Select("testUser", &jim)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("no jims response was empty")
	} else{
		fmt.Println("no jims:", jim)
	}

	jimmySend := testUser{
		Name: "Jimmy",
	}
	var jimmyReceive testUser
	ok, err = db.Create("testUser", &jimmySend, &jimmyReceive)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("jimmy response was empty")
	} else{
		fmt.Println("jimmy sent:", jimmySend)
		fmt.Println("jimmy received:", jimmyReceive)
	}

	jimmySend.ID = ""
	jimmySend.Name = "jimmy 2"
	ok, err = db.Update(jimmyReceive.ID, &jimmySend, &jimmyReceive)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("jimmy 2 response was empty")
	} else{
		fmt.Println("jimmy 2 sent:", jimmySend)
		fmt.Println("jimmy 2 received:", jimmyReceive)
	}

	var users []testUser
	ok, err = db.Query("select * from testUser", &users, nil)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("users response was empty")
	} else{
		fmt.Println("users:", users)
	}

	var user testUser
	ok, err = db.Query("select * from testUser where id = " + jimmyReceive.ID, &user, nil)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("users response was empty")
	} else{
		fmt.Println("user:", user)
	}

	var users2 []testUser
	ok, err = db.Query("select * from testUser where id = " + jimmyReceive.ID, &users2, nil)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("users response was empty")
	} else{
		fmt.Println("user in slice:", users2)
	}

	var user3 testUser
	ok, err = db.Query("select * from testUser where name = 'jimmy'", &user3, nil)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("users response was empty")
	} else{
		fmt.Println("user 3:", user3)
	}

	var user2 testUser
	ok, err = db.Query("selec t* from testUsr where name = 'jim'", &user2,nil)
	if err != nil{
		panic(err)
	}
	if !ok{
		fmt.Println("users response was empty")
	} else{
		fmt.Println("user 2:", user2)
	}
}
