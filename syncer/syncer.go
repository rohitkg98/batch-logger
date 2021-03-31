package syncer

import (
	"fmt"
	"net/http"
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

// Creates a function with provided payloads in scope
func CreateLogHandler(payloads *[]Log) echo.HandlerFunc {

	return func(ctx echo.Context) error {
		logEntry := new(Log)
		if err := ctx.Bind(logEntry); err != nil {
			ctx.Logger().Error("Invalid Request Body: ", ctx.Request().Body)
			return ctx.String(http.StatusBadRequest, "Invalid Request Body")
		}

		*payloads = append(*payloads, *logEntry)
		fmt.Printf("%v", payloads)
		return ctx.String(http.StatusOK, "OK")
	}
}

// Run as goroutine to run alongside server
// Periodically dumps payloads to provided endpoint, after provided interval
func StartIntervalSyncer(payloads *[]Log, batchInterval int) {
	for {
		time.Sleep(time.Second * time.Duration(batchInterval))
		println("Interval Syncer running")
		fmt.Printf("%v", payloads)
		println("Interval Syncer Complete")
	}
}
