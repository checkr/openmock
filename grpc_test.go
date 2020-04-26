package openmock

import (
	"context"
	"log"
	"testing"
	"time"

	pb "github.com/checkr/openmock/demo_protobuf"
	"google.golang.org/grpc"
)

const (
	grpcaddress = "localhost:50051"
)

// need to have a running grpc server on openmock to run this test
func Test(t *testing.T) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(grpcaddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewExampleServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.ExampleMethod(ctx, &pb.ExampleRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v, return %v", err, r)
	}
	log.Fatalf("Greeting: %d, %s", r.GetCode(), r.GetMessage())
}
