package server

import (
	"database/sql"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func ConfigServer(db *sql.DB) {
	router := fasthttprouter.New()


	router.POST("/", func(ctx *fasthttp.RequestCtx) {
		CheckAvailability(ctx, db)
	})

	fasthttp.ListenAndServe(":6061", fasthttp.CompressHandler(router.Handler))
}