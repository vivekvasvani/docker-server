package server

import (
	"database/sql"

	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

func ConfigServer(db *sql.DB) {
	router := fasthttprouter.New()

	//Check if servers are available
	router.GET("/checkavailability/:type", func(ctx *fasthttp.RequestCtx) {
		Checkavailability(ctx, db)
	})

	//Check current status of instances
	router.GET("/checkinstancestatus/:instanceid", func(ctx *fasthttp.RequestCtx) {
		CheckInstanceStatus(ctx, db)
	})

	//Start instances POST
	router.POST("/startinstances", func(ctx *fasthttp.RequestCtx) {
		CheckAndStartInstance(ctx, db)
	})

	//Start instances GET
	router.GET("/startinstances/:instanceid", func(ctx *fasthttp.RequestCtx) {
		StartInstancesGet(ctx, db)
	})

	//Stop instances DELETE
	router.DELETE("/stopinstances", func(ctx *fasthttp.RequestCtx) {
		StopInstancesWithPayload(ctx, db)
	})

	//Start Docker on environment
	router.POST("/playwithdocker", func(ctx *fasthttp.RequestCtx) {
		StartDockerEnvironment(ctx, db)
	})

	//Callback API
	router.PUT("/updatejobstatus/:jobid", func(ctx *fasthttp.RequestCtx) {
		UpdateJobStatusAndNotifyUser(ctx, db)
	})
	fmt.Println(fasthttp.ListenAndServe(":6002", fasthttp.CompressHandler(router.Handler)))
}
