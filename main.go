package main

import (
	"database/sql"
	"fmt"

	"github.com/vivekvasvani/docker-server/server"

	_ "github.com/go-sql-driver/mysql"
)

//global db variable
var db *sql.DB

func main() {
	getDB()
	server.ConfigServer(db)
}

func getDB() {
	var err error

	db, err = sql.Open("mysql", "abhijit:Myntra@123@tcp(192.168.20.59:3306)/nazgul")
	if err != nil {
		fmt.Print(err.Error())
		panic("Not able to Connect To DataBase")
	}
	fmt.Println(db)
	//defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Print("Error :", err)
	}
}
