package openmock

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/checkr/openmock/demo_protobuf"
)

// GRPCServiceMethodResponseMap Map of services, methods with the response protobuf they expect.
// This is needed to give a proper response to the GRPC client.
var GRPCServiceMethodResponseMap = map[string]map[string]RequestResponsePair{
	"demo_protobuf.ExampleService": {
		"ExampleMethod": RequestResponsePair{
			Request:  &demo_protobuf.ExampleRequest{},
			Response: &demo_protobuf.ExampleResponse{},
		},
	},
}

type RequestResponsePair struct {
	Request  proto.Message
	Response proto.Message
}

// convertBodyToJSON is how we support JSONPath to take values from GRPC requests and include them in responses
func convertBodyToJSON(h ExpectGRPC, body []byte) string {
	m := GRPCServiceMethodResponseMap[h.Service][h.Method].Request

	// first 5 bytes are compression and size information
	err := proto.Unmarshal(body[5:], m)

	if err != nil {
		logrus.Fatalf("error unmarshalling body and message: %v, %v", err, m)
	}

	jsonRequestMsg, err := protojson.Marshal(m)

	if err != nil {
		logrus.Fatalf("error marshalling proto to json %v", m)
	}

	return string(jsonRequestMsg)
}

func (om *OpenMock) startGRPC() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		logrus.WithFields(logrus.Fields{
			"grpc_req": string(reqBody),
			"grpc_res": string(resBody),
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
					JSONRequestBody := convertBodyToJSON(h, body)

					c := Context{
						GRPCContext: ec,
						GRPCHeader:  ec.Request().Header,
						GRPCPayload: JSONRequestBody,
						GRPCMethod:  h.Method,
						GRPCService: h.Service,
						om:          om,
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

	e.Logger.Info("Serving GRPC traffic on %s:%d", om.GRPCHost, om.GRPCPort)
	e.Logger.Fatal(
		e.Start(fmt.Sprintf("%s:%d", om.GRPCHost, om.GRPCPort)),
	)
}
