package openmock

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func (om *OpenMock) startGRPC() {
	logrus.Infof("GRPC Mock Not Implemented")

	fmt.Println("Starting GripMock in OpenMock")
	// Pulled from gripmock
	// https://github.com/tokopedia/gripmock

	if os.Getenv("GOPATH") == "" {
		log.Fatal("output is not provided and GOPATH is empty")
	}
	output := os.Getenv("GOPATH") + "/src/grpc"

	// for safety
	output += "/"
	if _, err := os.Stat(output); os.IsNotExist(err) {
		os.Mkdir(output, os.ModePerm)
	}

	// parse proto files
	protoPaths := []string{"protos/example.proto"}
	protos, err := parseProto(protoPaths)
	if err != nil {
		log.Fatal("can't parse proto ", err)
	}

	// generate pb.go using protoc
	generateProtoc(protoPaths, output)

	// generate grpc server based on proto
	generateGrpcServer(output, fmt.Sprintf("%s:%d",
		om.GRPCHost, om.GRPCPort), protos)

	// build the server
	buildServer(output, protoPaths)

	// and run
	run, runerr := runGrpcServer(output)

	var term = make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	select {
	case err := <-runerr:
		log.Fatal(err)
	case <-term:
		fmt.Println("Stopping gRPC Server")
		run.Process.Kill()
	}
}
