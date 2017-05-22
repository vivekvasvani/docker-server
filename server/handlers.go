package server

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"
)

func CheckAndStartEnvironment(ctx *fasthttp.RequestCtx, db *sql.DB) {
	var request Request
	err := json.Unmarshal(ctx.Request.Body(), request)
	checkError(err)
	fmt.Println(request)
	/*
			switch steps.EnvType {
			case "msg":
				msgIP = "192.168.0.190"
				ExecuteCommandsOnremote(msgIP, steps)
			case "devx":

			case "labs":
		}
	*/

}

func checkError(err error) {
	if err != nil {
		fmt.Print(err)
	}
}
