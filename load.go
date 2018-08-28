package openmock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Load returns a map of Mocks
func (om *OpenMock) Load() error {
	f, err := loadYAML(om.TemplatesDir)
	if err != nil {
		return err
	}
	mocks := []*Mock{}
	if err := yaml.UnmarshalStrict(f, &mocks); err != nil {
		return err
	}
	om.populateMockRepo(mocks)
	return nil
}

func (om *OpenMock) populateMockRepo(mocks []*Mock) {
	r := &MockRepo{
		HTTPMocks:  HTTPMocks{},
		KafkaMocks: KafkaMocks{},
		AMQPMocks:  AMQPMocks{},
	}
	for i := range mocks {
		m := mocks[i]
		m.loadFile(om.TemplatesDir)

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
}

func loadYAML(searchDir string) ([]byte, error) {
	w := &bytes.Buffer{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if f != nil && (strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml")) {
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
	return []byte(w.String()), nil
}

func (m *Mock) loadFile(baseDir string) {
	for i := range m.Actions {
		a := &m.Actions[i]
		if !structs.IsZero(a.ActionPublishAMQP) {
			amqp := &a.ActionPublishAMQP
			if amqp.PayloadFromFile != "" && amqp.Payload == "" {
				amqp.Payload = readFile(m.Key, baseDir, amqp.PayloadFromFile)
			}
		}
		if !structs.IsZero(a.ActionPublishKafka) {
			kafka := &a.ActionPublishKafka
			if kafka.PayloadFromFile != "" && kafka.Payload == "" {
				kafka.Payload = readFile(m.Key, baseDir, kafka.PayloadFromFile)
			}
		}
		if !structs.IsZero(a.ActionReplyHTTP) {
			h := &a.ActionReplyHTTP
			if h.BodyFromFile != "" && h.Body == "" {
				h.Body = readFile(m.Key, baseDir, h.BodyFromFile)
			}
		}
	}
	logrus.Infof("template with key:%s loaded.", m.Key)
}

func readFile(templateKey string, baseDir string, filePath string) string {
	path := fmt.Sprintf("%s/%s", baseDir, filePath)
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"template_key": templateKey,
			"err":          err,
			"path":         path,
		}).Errorf("failed to load file")
		return ""
	}
	return string(dat)

}
