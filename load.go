package openmock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structs"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const (
	redisTemplatesStore = "redis_templates_store"
)

// Load returns a map of Mocks
func (om *OpenMock) Load() error {
	f, err := loadYAML(om.TemplatesDir)
	if err != nil {
		return err
	}

	r, err := loadRedis(om.redis)
	if err != nil {
		return err
	}

	b := bytes.Join([][]byte{f, r}, []byte("\n"))

	mocks := []*Mock{}
	if err := yaml.UnmarshalStrict(b, &mocks); err != nil {
		return err
	}
	for _, m := range mocks {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	if err := om.populateTemplates(mocks); err != nil {
		return err
	}
	om.populateBehaviors(mocks)
	return nil
}

func (om *OpenMock) populateTemplates(mocks []*Mock) error {
	c := &Context{}
	for i := range mocks {
		m := mocks[i]
		if m.Kind == KindTemplate {
			_, err := c.Render(fmt.Sprintf(`{{define "%s"}}%s{{end}}`, m.Key, m.Template))
			if err != nil {
				return err
			}
			om.repo.Templates = append(om.repo.Templates, m)
		}
	}
	return nil
}

func (om *OpenMock) populateBehaviors(mocks []*Mock) {
	r := om.repo

	for i := range mocks {
		m := mocks[i]
		m.loadFile(om.TemplatesDir)
		if m.Include != "" {
			m = r.Behaviors[m.Include].patchedWith(*m)
		}
		r.Behaviors[m.Key] = m

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
}

func loadRedis(doer RedisDoer) ([]byte, error) {
	if doer == nil {
		return nil, nil
	}

	logrus.Infof("Start to load templates from redis")
	v, err := doer.Do("HGETALL", redisTemplatesStore)
	m, err := redis.StringMap(v, err)
	ss := []string{}
	for _, s := range m {
		ss = append(ss, s)
	}
	return []byte(strings.Join(ss, "\n")), nil
}

func loadYAML(searchDir string) ([]byte, error) {
	logrus.Infof("Start to load templates from: %s", searchDir)

	w := &bytes.Buffer{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

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
	logrus.Infof("Done with loading templates from: %s", searchDir)
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

func (m Mock) patchedWith(patch Mock) *Mock {
	baseStruct := structs.New(&m)
	patchStruct := structs.New(patch)
	for _, field := range patchStruct.Fields() {
		if !field.IsZero() {
			baseStruct.Field(field.Name()).Set(field.Value())
		}
	}
	return &m
}
