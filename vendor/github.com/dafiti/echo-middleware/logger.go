package middleware

import (
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"io"
)

// LoggerWithOutput returns a Logger middleware with output.
// See: `Logger()`.
func LoggerWithOutput(w io.Writer) echo.MiddlewareFunc {
	config := mw.DefaultLoggerConfig
	config.Output = w

	return mw.LoggerWithConfig(config)
}
