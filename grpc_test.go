package openmock

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/checkr/openmock/demo_protobuf"
)

// arbitraryPort returns a non-used port.
func arbitraryPort() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	defer l.Close()
	addr := l.Addr().String()
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	p, _ := strconv.Atoi(port)
	return p
}

var (
	testGRPCHost       = "127.0.0.1"
	testGRPCPort       = arbitraryPort()
	demoGRPCServiceMap = map[string]GRPCService{
		"demo_protobuf.ExampleService": {
			"ExampleMethod": GRPCRequestResponsePair{
				Request:  &demo_protobuf.ExampleRequest{},
				Response: &demo_protobuf.ExampleResponse{},
			},
		},
	}
)

// need to have a running grpc server on openmock to run this test
func TestGRPCServer(t *testing.T) {
	grpcaddress := fmt.Sprintf("%s:%d", testGRPCHost, testGRPCPort)
	om := &OpenMock{}
	om.SetupRepo()
	om.TemplatesDir = "./demo_templates"
	om.GRPCPort = testGRPCPort
	om.GRPCHost = testGRPCHost
	om.GRPCServiceMap = demoGRPCServiceMap
	err := om.Load()
	assert.NoError(t, err)
	go om.startGRPC()

	t.Run("happy code path", func(t *testing.T) {
		// Set up a connection to the server.
		conn, err := grpc.Dial(grpcaddress, grpc.WithInsecure(), grpc.WithBlock())
		defer conn.Close()
		assert.NoError(t, err)

		c := demo_protobuf.NewExampleServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := c.ExampleMethod(ctx, &demo_protobuf.ExampleRequest{Two: "success"})
		assert.NoError(t, err)
		assert.Equal(t, 227, int(r.GetCode()))
		assert.Equal(t, "success", r.GetMessage())
	})

	t.Run("ok with condition", func(t *testing.T) {
		// Set up a connection to the server.
		conn, err := grpc.Dial(grpcaddress, grpc.WithInsecure(), grpc.WithBlock())
		defer conn.Close()
		assert.NoError(t, err)

		c := demo_protobuf.NewExampleServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := c.ExampleMethod(ctx, &demo_protobuf.ExampleRequest{Two: "zzzz"})
		assert.NoError(t, err)
		assert.Equal(t, 200, int(r.GetCode()))
		assert.Equal(t, "yyyy", r.GetMessage())
	})
}
