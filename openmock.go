package openmock

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// OpenMock holds all the configuration of running openmock
type OpenMock struct {
	// HTTP host and port to serve http mocks
	HTTPEnabled bool
	HTTPPort    int
	HTTPHost    string

	// Kafka Settings
	KafkaEnabled             bool
	KafkaClientID            string
	KafkaSeedBrokers         []string
	KafkaConsumePipelineFunc KafkaPipelineFunc
	KafkaPublishPipelineFunc KafkaPipelineFunc

	// AMQP URL
	AMQPEnabled bool
	AMQPURL     string

	// The templates directory to load the YAML files
	// The dir is relative to the runtime binary
	TemplatesDir string

	// Prviates
	repo        *MockRepo
	kafkaClient *kafkaClient
}

// Start starts the openmock
func (om *OpenMock) Start() {
	err := om.Load()
	if err != nil {
		logrus.Fatalf("%s: %s", "failed to load yaml templates for mocks", err)
	}

	if om.HTTPEnabled {
		go om.startHTTP()
	}
	if om.KafkaEnabled {
		go om.startKafka()
	}
	if om.AMQPEnabled {
		go om.startAMQP()
	}

	select {}
}

// Load returns a map of Mocks
func (om *OpenMock) Load() error {
	f, err := loadFiles(om.TemplatesDir)
	if err != nil {
		return err
	}
	mocks := []*Mock{}
	if err := yaml.UnmarshalStrict(f, &mocks); err != nil {
		return err
	}
	r := &MockRepo{
		HTTPMocks:  HTTPMocks{},
		KafkaMocks: KafkaMocks{},
		AMQPMocks:  AMQPMocks{},
	}
	for i := range mocks {
		m := mocks[i]

		if !structs.IsZero(m.Expect.HTTP) {
			_, ok := r.HTTPMocks[m.Expect.HTTP]
			if !ok {
				r.HTTPMocks[m.Expect.HTTP] = []*Mock{m}
			} else {
				r.HTTPMocks[m.Expect.HTTP] = append(r.HTTPMocks[m.Expect.HTTP], m)
			}
		}
		if !structs.IsZero(m.Expect.Kafka) {
			_, ok := r.KafkaMocks[m.Expect.Kafka]
			if !ok {
				r.KafkaMocks[m.Expect.Kafka] = []*Mock{m}
			} else {
				r.KafkaMocks[m.Expect.Kafka] = append(r.KafkaMocks[m.Expect.Kafka], m)
			}
		}
		if !structs.IsZero(m.Expect.AMQP) {
			_, ok := r.AMQPMocks[m.Expect.AMQP]
			if !ok {
				r.AMQPMocks[m.Expect.AMQP] = []*Mock{m}
			} else {
				r.AMQPMocks[m.Expect.AMQP] = append(r.AMQPMocks[m.Expect.AMQP], m)
			}
		}
	}
	om.repo = r
	return nil
}

func loadFiles(searchDir string) ([]byte, error) {
	w := &bytes.Buffer{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			w.Write(content)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return []byte(os.ExpandEnv(w.String())), nil
}
