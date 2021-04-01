package main

import (
	"batch-logger/syncer"
	"batch-logger/utils"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Globals struct {
	payloads syncer.Payloads
}

var (
	batchSize     = utils.GetEnvAsInt("BATCH_SIZE")
	batchInterval = utils.GetEnvAsInt("BATCH_INTERVAL")
	globals       = new(Globals)
)

func main() {
	globals.payloads = syncer.Payloads{
		BatchSize: batchSize,
		Stream:    make(chan syncer.ProcessOrLog),
	}
	// Echo instance
	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "Request received: method=${method}, uri=${uri}, duration=${latency_human}\n",
	}))
	e.Use(middleware.Recover())

	// Routes
	e.GET(
		"/healthz",
		func(ctx echo.Context) error {
			return ctx.String(http.StatusOK, "OK")
		},
	)

	s := syncer.CreateSyncer(batchSize, batchInterval, globals.payloads, e)

	e.POST("/log", s.CreateLogHandler())

	// Start gorountine
	go s.SyncAtIntervals()
	go s.Listen()

	// Start server
	e.Logger.SetOutput(os.Stdout)
	e.Logger.SetLevel(log.INFO)
	e.Logger.Info("Server started!")
	e.Logger.Fatal(e.Start(":8080"))
}
