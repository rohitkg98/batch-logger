package syncer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
)

// Expected Request Data Format
type Log struct {
	UserId int     `json:"user_id"`
	Total  float64 `json:"total"`
	Title  string  `json:"title"`
	Meta   struct {
		Logins []struct {
			Time string `json:"time"`
			Ip   string `json:"ip"`
		} `json:"logins"`
		PhoneNumbers struct {
			Home   string `json:"home"`
			Mobile string `json:"mobile"`
		} `json:"phone_numbers"`
	} `json:"meta"`
	Completed bool `json:"completed"`
}

type Syncer struct {
	batchSize     int
	batchInterval int
	payloads      Payloads
	e             *echo.Echo
}

func CreateSyncer(batchSize int, batchInterval int, payloads Payloads, e *echo.Echo) Syncer {
	return Syncer{
		batchSize,
		batchInterval,
		payloads,
		e,
	}
}

// Create a Handler route for receiving Log Entries
func (s Syncer) CreateLogHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		logEntry := new(Log)
		if err := ctx.Bind(logEntry); err != nil {
			ctx.Logger().Error("Invalid Request Body: ", ctx.Request().Body)
			return ctx.String(http.StatusBadRequest, "Invalid Request Body")
		}

		s.payloads.Add(*logEntry)

		return ctx.String(http.StatusOK, "OK")
	}
}

// Run as goroutine to run alongside server
// Periodically triggers sync process
func (s Syncer) SyncAtIntervals() {
	for {
		time.Sleep(time.Second * time.Duration(s.batchInterval))
		s.payloads.Process()
	}
}

// Listens to payload stream
// if Process is triggered, syncs to POST_ENDPOINT
// also syncs when Batch Size has reached
func (s Syncer) Listen() {
	payloads := make([]Log, 0, s.batchSize)
	for {
		payload := <-s.payloads.Stream

		if !payload.process {
			payloads = append(payloads, payload.entry)
		}

		if (payload.process || len(payloads) >= s.batchSize) && len(payloads) > 0 {
			// Call Sync as a goroutine so it executes concurrently
			go syncToPostEndpoint(payloads, s.e)
			payloads = make([]Log, 0, s.batchSize)
		}
	}
}

func readPostEndpoint() string {
	env, exists := os.LookupEnv("POST_ENDPOINT")

	if !exists {
		panic("Please set Environment Variable: POST_ENDPOINT")
	}
	return env
}

var postEndpoint = readPostEndpoint()

type Payload struct {
	Payloads []Log `json:"payloads"`
}

func syncToPostEndpoint(payloads []Log, e *echo.Echo) {
	jsonData, err := json.Marshal(Payload{
		Payloads: payloads,
	})

	if err != nil {
		panic("Payloads cannot be Marshalled\n")
	}

	var statusCode int
	start := time.Now()
	for count := 0; count < 3; count++ {
		resp, err := http.Post(postEndpoint, "application/json", bytes.NewBuffer(jsonData))

		if err == nil {
			statusCode = resp.StatusCode
			break
		}

		// Error non-nil and three requests have been made.
		if count == 2 {
			panic(fmt.Sprintf("Calls to %s failed three times in a row.", postEndpoint))
		}
		// Wait 2 seconds before retrying.
		time.Sleep(2 * time.Second)
	}
	stop := time.Now()

	// Log BatchSize, Response Code and Duration on success
	e.Logger.Infof("BatchSize: %d, RespStatusCode: %d, Duration: %s",
		len(payloads),
		statusCode,
		stop.Sub(start),
	)
}
