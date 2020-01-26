// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	errors "github.com/go-openapi/errors"

	"github.com/checkr/openmock"
	"github.com/checkr/openmock/pkg/admin"
	"github.com/checkr/openmock/swagger_gen/restapi/operations"
	"github.com/sirupsen/logrus"

	"io/ioutil"

	yaml "github.com/ghodss/yaml"

	"github.com/go-openapi/runtime"
)

//go:generate swagger generate server --target ../../swagger_gen --name OpenMock --spec ../../docs/api_docs/bundle.yaml

func configureFlags(api *operations.OpenMockAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.OpenMockAPI, customOpenmock *openmock.OpenMock) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf
	api.Logger = logrus.Infof
	api.ServerShutdown = admin.ServerShutdown

	api.YamlConsumer = MyYAMLConsumer()
	api.YamlProducer = MyYAMLProducer()

	api.JSONConsumer = runtime.JSONConsumer()
	api.JSONProducer = runtime.JSONProducer()

	shouldRunAdmin := admin.Setup(api, customOpenmock)
	if shouldRunAdmin {
		return setupGlobalMiddleware(api.Serve(setupMiddlewares))
	}

	logrus.Info("Admin API disabled")
	waitForSignal()
	panic("terminating")
}

func waitForSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
	)
	<-signalChan
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {

}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return admin.SetupGlobalMiddleware(handler)
}

// custom YAML consumer using github.com/ghodss/yaml library. Using the one from
// yamlpc I ran into a weird bug where fields with _ in them would drop the underscore;
// e.g. field reply_http would become replyhttp, and be incompatible with
// original Openmock Mock's yaml definition.
// ditto for YAMLProducer
//
// this also seemed to fix an error that it couldn't handle fields that use
// additionalProperties:true in openapi, like reply_http's headers.
//
func MyYAMLConsumer() runtime.Consumer {
	return runtime.ConsumerFunc(func(r io.Reader, v interface{}) error {
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(buf, v)
	})
}

func MyYAMLProducer() runtime.Producer {
	return runtime.ProducerFunc(func(w io.Writer, v interface{}) error {
		b, _ := yaml.Marshal(v) // can't make this error come up
		_, err := w.Write(b)
		return err
	})
}
