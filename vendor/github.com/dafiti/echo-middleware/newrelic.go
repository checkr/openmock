package middleware

import (
	"fmt"
	"github.com/labstack/echo"
	nr "github.com/newrelic/go-agent"
)

const (
	// NEWRELIC_TXN defines the context key used to save newrelic transaction
	NEWRELIC_TXN = "newrelic-txn"
)

// NewRelic returns a middleware that collect request data for NewRelic
func NewRelic(appName string, licenseKey string) echo.MiddlewareFunc {
	config := nr.NewConfig(appName, licenseKey)
	app, err := nr.NewApplication(config)

	if err != nil {
		panic(fmt.Errorf("New relic: %s", err))
	}

	return NewRelicWithApplication(app)
}

// NewRelicWithApplication returns a NewRelic middleware with application.
// See: `NewRelic()`.
func NewRelicWithApplication(app nr.Application) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			transactionName := fmt.Sprintf("%s [%s]", c.Path(), c.Request().Method)
			txn := app.StartTransaction(transactionName, c.Response().Writer, c.Request())
			defer txn.End()

			c.Set(NEWRELIC_TXN, txn)

			err := next(c)

			if err != nil {
				txn.NoticeError(err)
			}

			return err
		}
	}
}
