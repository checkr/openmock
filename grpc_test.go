package openmock

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	pb "github.com/checkr/openmock/demo_protobuf"
)

const (
	grpcaddress = "localhost:50051"
)

// need to have a running grpc server on openmock to run this test
func TestGRPCServer(t *testing.T) {
	om := &OpenMock{}

	om.SetupRepo()

	om.TemplatesDir = "./demo_templates"
	om.GRPCPort = 50051
	err := om.Load()
	if err != nil {
		logrus.Fatalf("%s: %s", "failed to load yaml templates for mocks", err)
	}
	go om.startGRPC()

	// Set up a connection to the server.
	conn, err := grpc.Dial(grpcaddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewExampleServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.ExampleMethod(ctx, &pb.ExampleRequest{Two: "success"})
	if err != nil {
		log.Fatalf("could not greet: %v, return %v", err, r)
	}
	assert.Equal(t, 227, int(r.GetCode()))
	assert.Equal(t, "success", r.GetMessage())
}
