package openmock

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

var localFuncMap = template.FuncMap{
	"jsonPath": JSONPath,
	"xmlPath":  XMLPath,
	"uuidv5":   uuidv5,
}

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

	om *OpenMock
}

// cleanup replaces all the linebreaks and tabs with spaces
func cleanup(raw string) string {
	re := regexp.MustCompile(`\r?\n|\t`)
	return re.ReplaceAllString(raw, " ")
}

// Render renders the raw given the context
func (c *Context) Render(raw string) (string, error) {
	tmpl, err := template.New("").
		Funcs(sprig.TxtFuncMap()). // supported functions https://github.com/Masterminds/sprig/blob/master/functions.go
		Funcs(localFuncMap).
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
func (c *Context) MatchCondition(condition string) (r bool) {
	defer func() {
		logrus.WithFields(logrus.Fields{
			"HTTPBody":     c.HTTPBody,
			"KafkaPayload": c.KafkaPayload,
			"AMQPPayload":  c.AMQPPayload,
			"condition":    condition,
			"result":       r,
		}).Info("running MatchCondition")
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

// JSONPath uses xpath to find the info from JSON string
func JSONPath(expr string, tmpl string) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"tmpl": tmpl,
			"expr": expr,
		}).Info("running json xpath")
	}()

	if tmpl == "" {
		return "", nil
	}

	doc, err := jsonquery.Parse(strings.NewReader(tmpl))
	if err != nil {
		return "", err
	}
	node := jsonquery.FindOne(doc, expr)
	if node != nil {
		return node.InnerText(), nil
	}
	return "", nil
}

// XMLPath uses xpath to find the info from XML string
func XMLPath(expr string, tmpl string) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"tmpl": tmpl,
			"expr": expr,
		}).Info("running xml xpath")
	}()

	if tmpl == "" {
		return "", nil
	}

	doc, err := xmlquery.Parse(strings.NewReader(tmpl))
	if err != nil {
		return "", err
	}
	node := xmlquery.FindOne(doc, expr)
	if node != nil {
		return node.InnerText(), nil
	}
	return "", nil
}

func uuidv5(dat string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(dat)).String()
}
