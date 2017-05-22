package main

import (
	"database/sql"
	"fmt"

	"github.com/vivekvasvani/docker-server/server"
	utils "github.com/vivekvasvani/docker-server/utils"

	_ "github.com/go-sql-driver/mysql"
)

//global db variable
var db *sql.DB

func main() {
	getDB()
	utils.ExecuteCommandOnLocal("pwd")
	utils.ExecuteCommandOnLocal("sh remoteshell/messaging.sh")
	server.ConfigServer(db)
}

func getDB() {
	var err error

	db, err = sql.Open("mysql", "root:hike@tcp(10.128.20.71:3306)/docker")
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
