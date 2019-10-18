package openmock

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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

	om.populateMocks(f, false)
	if err != nil {
		return err
	}

	r, err := loadRedis(om.redis)
	if err != nil {
		return err
	}

	om.populateMocks(r, true)
	if err != nil {
		return err
	}
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

func (om *OpenMock) populateMocks(rawMocks []byte, verifyUniqueBehaviors bool) error {
	mocks := []*Mock{}
	if err := yaml.UnmarshalStrict(rawMocks, &mocks); err != nil {
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
	om.populateBehaviors(mocks, verifyUniqueBehaviors)
	return nil
}

func (om *OpenMock) populateBehaviors(mocks []*Mock, verifyUniqueBehaviors bool) error {
	r := om.repo

	for i := range mocks {
		m := mocks[i]
		m.loadFile(om.TemplatesDir)
		if r.Behaviors[m.Key] != nil {
			if verifyUniqueBehaviors {
				logrus.WithField("key", m.Key).Error("Overriding existing behavior")
				return errors.New("Overriding existing behavior from yaml not allowed")
			} else {
				logrus.WithField("key", m.Key).Info("Overriding existing behavior")
			}
		}
		r.Behaviors[m.Key] = m
	}

	for _, m := range r.Behaviors {
		if m.Kind == KindAbstractBehavior {
			continue
		}
		if m.Extend != "" {
			m = r.Behaviors[m.Extend].patchedWith(*m)
			r.Behaviors[m.Key] = m
		}
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

		if len(r.Behaviors[m.Key].Actions) > 0 {
			orderedActions := r.Behaviors[m.Key].Actions
			sort.Slice(orderedActions, func(i, j int) bool {
				return orderedActions[i].Order < orderedActions[j].Order
			})
			if !actionsEqual(orderedActions, r.Behaviors[m.Key].Actions) {
				m = r.Behaviors[m.Key].patchedWith(Mock{
					Actions: orderedActions,
				})
				r.Behaviors[m.Key] = m
			}
		}
	}

	return nil
}

func actionsEqual(lhs []ActionDispatcher, rhs []ActionDispatcher) bool {
	lhsString, lhsError := yaml.Marshal(lhs)
	rhsString, rhsError := yaml.Marshal(rhs)
	return !(lhsError != nil || rhsError != nil) && (string(lhsString) == string(rhsString))
}

func loadRedis(doer RedisDoer) (b []byte, err error) {
	if doer == nil {
		return nil, nil
	}

	logrus.Infof("Start to load templates from redis")
	v, err := doer.Do("HGETALL", redisTemplatesStore)
	m, err := redis.StringMap(v, err)
	if err != nil {
		return nil, err
	}
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
	return w.Bytes(), nil
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
	var values = make(map[string]interface{})
	for key, value := range m.Values {
		values[key] = value
	}
	for key, value := range patch.Values {
		values[key] = value
	}

	actions := append(m.Actions, patch.Actions...)

	baseStruct := structs.New(&m)
	patchStruct := structs.New(patch)
	for _, field := range patchStruct.Fields() {
		if !field.IsZero() {
			err := baseStruct.Field(field.Name()).Set(field.Value())
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"template_key": m.Key,
					"err":          err,
				}).Errorf("failed to extend")
				return nil
			}
		}
	}
	m.Values = values
	m.Actions = actions

	return &m
}
