package openmock

// Pulled from gripmock
// https://github.com/tokopedia/gripmock

import (
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
)

func TestGenerateServerFromProto(t *testing.T) {
	services, err := GetServicesFromProto(protofile)
	assert.NoError(t, err)
	err = GenerateServer(services, &Options{
		writer: os.Stdout,
	})
	assert.NoError(t, err)
}
