package main

import (
	"batch-logger/syncer"
	"batch-logger/utils"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Globals struct {
	Payloads []syncer.Log
}

var (
	batchSize     = utils.GetEnvAsInt("BATCH_SIZE")
	batchInterval = utils.GetEnvAsInt("BATCH_INTERVAL")
	logFile, _    = os.Open("./server.log")
	globals       = new(Globals)
)

func main() {
	globals.Payloads = make([]syncer.Log, 0, batchSize)
	// Echo instance
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, latency=${latency}\n",
		Output: logFile,
	}))
	e.Use(middleware.Recover())

	// Routes
	e.GET(
		"/healthz",
		func(ctx echo.Context) error {
			return ctx.String(http.StatusOK, "OK")
		},
	)

	e.POST("/log", syncer.CreateLogHandler(&globals.Payloads))

	// Start gorountine
	go syncer.StartIntervalSyncer(&globals.Payloads, batchInterval)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
