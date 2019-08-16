package openmock

import (
	"bytes"
	"net/http"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

// Context represents the context of the mock expectation
type Context struct {
	HTTPContext     echo.Context
	HTTPHeader      http.Header
	HTTPBody        string
	HTTPPath        string
	HTTPQueryString string

	KafkaTopic   string
	KafkaPayload string

	AMQPExchange   string
	AMQPRoutingKey string
	AMQPQueue      string
	AMQPPayload    string

	Values map[string]interface{}

	om *OpenMock
}

var globalTemplate = template.New("__global__")

// cleanup replaces all the linebreaks and tabs with spaces
func cleanup(raw string) string {
	re := regexp.MustCompile(`\r?\n|\t`)
	return re.ReplaceAllString(raw, " ")
}

// Render renders the raw given the context
func (c Context) Render(raw string) (out string, err error) {
	tmpl, err := globalTemplate.New("").
		Funcs(sprig.TxtFuncMap()). // supported functions https://github.com/Masterminds/sprig/blob/master/functions.go
		Funcs(genLocalFuncMap(c.om)).
		Option("missingkey=error").
		Parse(cleanup(raw))
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, c); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MatchCondition checks the condition given the context
func (c Context) MatchCondition(condition string) (r bool) {
	defer func() {
		if r {
			logrus.WithFields(logrus.Fields{
				"HTTPHeader":   c.HTTPHeader,
				"HTTPBody":     c.HTTPBody,
				"KafkaPayload": c.KafkaPayload,
				"AMQPPayload":  c.AMQPPayload,
				"condition":    condition,
				"result":       r,
			}).Debug("running MatchCondition")
		}
	}()

	if condition == "" {
		return true
	}

	result, err := c.Render(condition)
	if err != nil {
		logrus.WithField("err", err).Errorf("failed to render condition: %s", condition)
		return false
	}
	return result == "true"
}
