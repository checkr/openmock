package openmock

// Pulled from gripmock
// https://github.com/tokopedia/gripmock

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func getProtoName(path string) string {
	paths := strings.Split(path, "/")
	filename := paths[len(paths)-1]
	return strings.Split(filename, ".")[0]
}

func parseProto(protoPaths []string) ([]Service, error) {
	services := []Service{}
	for _, path := range protoPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			log.Fatal(fmt.Sprintf("Proto file '%s' not found", protoPaths))
		}
		byt, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatal("Error on reading proto " + err.Error())
		}
		service, err := GetServicesFromProto(string(byt))
		if err != nil {
			return nil, err
		}
		services = append(services, service...)
	}
	return services, nil
}

func generateProtoc(protoPaths []string, outputDir string) {
	protodirs := strings.Split(protoPaths[0], "/")
	protodir := ""
	if len(protodirs) > 0 {
		protodir = strings.Join(protodirs[:len(protodirs)-1], "/") + "/"
	}

	args := []string{"-I", protodir}
	args = append(args, protoPaths...)
	args = append(args, "--go_out=plugins=grpc:"+outputDir)
	protoc := exec.Command("protoc", args...)
	protoc.Stdout = os.Stdout
	protoc.Stderr = os.Stderr
	err := protoc.Run()
	if err != nil {
		log.Fatal("Fail on protoc ", err)
	}

	// change package to "main" on generated code
	for _, proto := range protoPaths {
		protoname := getProtoName(proto)
		sed := exec.Command("sed", `s/^package \w*$/package main/`, outputDir+protoname+".pb.go")
		sed.Stderr = os.Stderr
		sed.Stdout = os.Stdout
		err = sed.Run()
		if err != nil {
			log.Fatal("Fail on sed")
		}
	}
}

func generateGrpcServer(output, grpcAddr string, services []Service) {
	file, err := os.Create(output + "server.go")
	if err != nil {
		log.Fatal(err)
	}
	GenerateServer(services, &Options{
		writer:   file,
		grpcAddr: grpcAddr,
	})

}

func buildServer(output string, protoPaths []string) {
	args := []string{"build", "-o", output + "grpcserver", output + "server.go"}
	for _, path := range protoPaths {
		args = append(args, output+getProtoName(path)+".pb.go")
	}
	build := exec.Command("go", args...)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	err := build.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func runGrpcServer(output string) (*exec.Cmd, <-chan error) {
	run := exec.Command(output + "grpcserver")
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	err := run.Start()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("grpc server pid: %d\n", run.Process.Pid)
	runerr := make(chan error)
	go func() {
		runerr <- run.Wait()
	}()
	return run, runerr
}
