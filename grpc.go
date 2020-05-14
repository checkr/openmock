package openmock

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	grpcPayloadLen = 1
	grpcSizeLen    = 4
	grpcHeaderLen  = grpcPayloadLen + grpcSizeLen
)

// length-prefixed message, see https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md
func msgHeader(data []byte) (hdr []byte, payload []byte) {
	hdr = make([]byte, grpcHeaderLen)

	hdr[0] = byte(0)

	// Write length of payload into buf
	binary.BigEndian.PutUint32(hdr[grpcPayloadLen:], uint32(len(data)))
	return hdr, data
}

// GRPCService is a map of service_name => GRPCRequestResponsePair
type GRPCService map[string]GRPCRequestResponsePair

// GRPCRequestResponsePair is a pair of proto.Message to define
// the message schema of request and response of a method
type GRPCRequestResponsePair struct {
	Request  proto.Message
	Response proto.Message
}

func (om *OpenMock) convertJSONToH2Response(ctx Context, resJSON string) (header []byte, data []byte, err error) {
	if om.GRPCServiceMap == nil {
		return nil, nil, fmt.Errorf("empty GRPCServiceMap")
	}

	if _, ok := om.GRPCServiceMap[ctx.GRPCService]; !ok {
		return nil, nil, fmt.Errorf("invalid service in GRPCServiceMap. %s", ctx.GRPCService)
	}

	if _, ok := om.GRPCServiceMap[ctx.GRPCService][ctx.GRPCMethod]; !ok {
		return nil, nil, fmt.Errorf("invalid method in GRPCServiceMap[%s]. %+v", ctx.GRPCService, ctx.GRPCMethod)
	}

	res := om.GRPCServiceMap[ctx.GRPCService][ctx.GRPCMethod].Response
	err = protojson.Unmarshal([]byte(resJSON), res)
	if err != nil {
		return nil, nil, err
	}
	b, err := proto.Marshal(res)
	if err != nil {
		return nil, nil, err
	}

	header, data = msgHeader(b)
	return header, data, nil
}

// convertRequestBodyToJSON is how we support JSONPath to take values from GRPC requests and include them in responses
func (om *OpenMock) convertRequestBodyToJSON(h ExpectGRPC, body []byte) (string, error) {
	if om.GRPCServiceMap == nil {
		return "", fmt.Errorf("empty GRPCServiceMap")
	}

	if _, ok := om.GRPCServiceMap[h.Service]; !ok {
		return "", fmt.Errorf("invalid service in GRPCServiceMap. %s", h.Service)
	}

	if _, ok := om.GRPCServiceMap[h.Service][h.Method]; !ok {
		return "", fmt.Errorf("invalid method in GRPCServiceMap[%s]. %+v", h.Service, h.Method)
	}

	req := om.GRPCServiceMap[h.Service][h.Method].Request

	if len(body) <= grpcSizeLen {
		return "", fmt.Errorf("invalid grpc body length. length: %d", len(body))
	}
	// first grpcSizeLen bytes are compression and size information
	err := proto.Unmarshal(body[(grpcSizeLen+1):], req)

	if err != nil {
		return "", err
	}

	jsonRequestMsg, err := protojson.Marshal(req)

	if err != nil {
		return "", err
	}

	return string(jsonRequestMsg), nil
}

func (om *OpenMock) prepareGRPCEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		logrus.WithFields(logrus.Fields{
			"grpc_path":   c.Path(),
			"grpc_method": c.Request().Method,
			"grpc_host":   c.Request().Host,
			"grpc_req":    string(reqBody),
			"grpc_res":    string(resBody),
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
					JSONRequestBody, err := om.convertRequestBodyToJSON(h, body)
					if err != nil {
						logrus.WithError(err).Error("failed to convert gRPC request body to JSON")
						return err
					}

					c := Context{
						GRPCContext: ec,
						GRPCHeader:  ec.Request().Header,
						GRPCPayload: JSONRequestBody,
						GRPCMethod:  h.Method,
						GRPCService: h.Service,
						om:          om,
					}

					ms.DoActions(c)
					return nil
				},
			)
		}(h, ms)
	}
	return e
}

func (om *OpenMock) startGRPC() {
	s := &http2.Server{
		MaxConcurrentStreams: 250,
		MaxReadFrameSize:     1048576,
		IdleTimeout:          10 * time.Second,
	}
	e := om.prepareGRPCEcho()
	e.Logger.Fatal(e.StartH2CServer(fmt.Sprintf("%s:%d", om.GRPCHost, om.GRPCPort), s))
}
