package openmock

import (
	"fmt"
	"github.com/checkr/openmock/demo_protobuf"
	"io/ioutil"
	"time"

	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/http2"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/runtime/protoiface"
)

// Map of services, methods with the response protobuf they expect.
// This is needed to give a proper response to the GRPC client.
var ServiceMethodResponseMap = map[string]map[string]protoiface.MessageV1{
	"demo_protobuf.ExampleService":
		{
		"ExampleMethod": &demo_protobuf.ExampleResponse{},
		},
}

func (om *OpenMock) startGRPC() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		logrus.WithFields(logrus.Fields{
			"http_req": string(reqBody),
			"http_res": string(resBody),
		}).Info()
	}))
	if om.CorsEnabled {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowCredentials: true,
			AllowHeaders:     []string{"*"},
			AllowMethods:     []string{"*"},
		}))
	}
	mocks := om.repo.GRPCMocks
	for h, ms := range mocks {
		func(h ExpectGRPC, ms MocksArray) {
			e.Match(
				[]string{"POST"},
				fmt.Sprintf("/%s/%s", h.Service, h.Method),
				func(ec echo.Context) error {
					body, _ := ioutil.ReadAll(ec.Request().Body)
					c := Context{
						GRPCContext:     ec,
						GRPCHeader:      ec.Request().Header,
						GRPCPayload:     string(body),
						GRPCMethod:      h.Method,
						GRPCService:     h.Service,
						om:              om,
					}

					return ms.DoActions(c)
				},
			)
		}(h, ms)
	}

	s := &http2.Server{
		MaxConcurrentStreams: 250,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	e.Logger.Fatal(e.StartH2CServer(fmt.Sprintf("%s:%d", om.GRPCHost, om.GRPCPort), s))

	e.Logger.Fatal(
		e.Start(fmt.Sprintf("%s:%d", om.GRPCHost, om.GRPCPort)),
	)
}
