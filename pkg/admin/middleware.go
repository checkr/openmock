package admin

import "net/http"

// ServerShutdown is a callback function that will be called when
// we tear down the openmock server
func ServerShutdown() {
}

// SetupGlobalMiddleware setup the global middleware
func SetupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
