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

// Creates a function with provided payloads and batchSize in scope
// Another way to do this would be create these as a method on the Globals struct
func CreateLogHandler(payloads *[]Log, batchSize int) echo.HandlerFunc {

	return func(ctx echo.Context) error {
		logEntry := new(Log)
		if err := ctx.Bind(logEntry); err != nil {
			ctx.Logger().Error("Invalid Request Body: ", ctx.Request().Body)
			return ctx.String(http.StatusBadRequest, "Invalid Request Body")
		}

		*payloads = append(*payloads, *logEntry)

		if len(*payloads) >= batchSize {
			// Sync can be called as a goroutine
			// But doing so can result in payload size increase before sync
			syncToPostEndpoint(payloads, ctx.Echo())
		}

		return ctx.String(http.StatusOK, "OK")
	}
}

// Run as goroutine to run alongside server
// Periodically dumps payloads to provided endpoint, after provided interval
func StartIntervalSyncer(payloads *[]Log, batchInterval int, e *echo.Echo) {
	for {
		time.Sleep(time.Second * time.Duration(batchInterval))
		if len(*payloads) > 0 {
			syncToPostEndpoint(payloads, e)
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

func syncToPostEndpoint(payloads *[]Log, e *echo.Echo) {
	jsonData, err := json.Marshal(Payload{
		Payloads: *payloads,
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
		// This currently blocks the API, considering we don't want new entries
		// in the in-memory cache because that will exceed batch size.
		// To unblock, we can call sync as a go routine.
		time.Sleep(2 * time.Second)
	}
	stop := time.Now()

	// Log BatchSize, Response Code and Duration on success
	e.Logger.Infof("BatchSize: %d, RespStatusCode: %d, Duration: %s",
		len(*payloads),
		statusCode,
		stop.Sub(start),
	)
	// Clear the in-memory cache
	*payloads = make([]Log, 0, len(*payloads))
}
