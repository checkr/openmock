package openmock

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGripmockHelpers(t *testing.T) {
	t.Run("getProtoName returns filename", func(t *testing.T) {
		path := "path/to/filename.json"
		filename := getProtoName(path)
		assert.Equal(t, "filename", filename)
	})

	t.Run("parseProto service", func(t *testing.T) {
		paths := []string{"protos/example.proto"}
		services, err := parseProto(paths)
		assert.Equal(t, nil, err)
		assert.Equal(t, "MyService", services[0].Name)
	})

	// This has side effects and generates a file..
	// TODO: mock the things called here!!
	t.Run("generateProtoc does generate pb.go file", func(t *testing.T) {
		paths := []string{"protos/example.proto"}
		outputDir := "./temp/grpc/"
		exec.Command("mkdir -p " + outputDir)
		generateProtoc(paths, outputDir)
	})

	t.Run("generateGrpcServer does generate the server", func(t *testing.T) {
		paths := []string{"protos/example.proto"}
		services, _ := parseProto(paths)
		outputDir := "./temp/grpc/"
		generateGrpcServer(outputDir, "0.0.0.0:9997", services)
	})

	// t.Run("buildServer", func(t *testing.T) {
	// 	outputDir := "./temp/grpc/"
	// 	paths := []string{"protos/example.proto"}
	// 	buildServer(outputDir, paths)
	// })
}
