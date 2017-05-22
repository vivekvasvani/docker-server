package server

import (
	"database/sql"

	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func ConfigServer(db *sql.DB) {
	router := fasthttprouter.New()

	router.POST("/startenvironment", func(ctx *fasthttp.RequestCtx) {
		CheckAndStartEnvironment(ctx, db)
	})

	fmt.Println(fasthttp.ListenAndServe(":6061", fasthttp.CompressHandler(router.Handler)))

}
