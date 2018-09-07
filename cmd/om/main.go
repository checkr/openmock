package main

import (
	"github.com/checkr/openmock"
)

func main() {
	om := &openmock.OpenMock{}
	om.ParseEnv()
	om.Start()
}
