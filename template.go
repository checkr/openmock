package openmock

import (
	"bytes"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

const (
	funcJSONXpath   = "json_xpath"
	funcXMLXpath    = "xml_xpath"
	funcMatchString = "match_string"
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

	om *OpenMock
}

func removeLinebreaks(raw string) string {
	re := regexp.MustCompile(`\r?\n`)
	return re.ReplaceAllString(raw, "")
}

// Render renders the raw given the context
func (c *Context) Render(raw string) (string, error) {
	tmpl, err := template.
		New("").
		Funcs(template.FuncMap{
			funcJSONXpath:   JSONXPath,
			funcXMLXpath:    XMLXPath,
			funcMatchString: regexp.MatchString,
		}).Parse(removeLinebreaks(raw))
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

// JSONXPath uses xpath to find the info from JSON string
func JSONXPath(tmpl string, expr string) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"tmpl": tmpl,
			"expr": expr,
		}).Info("running json_xpath")
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

// XMLXPath uses xpath to find the info from XML string
func XMLXPath(tmpl string, expr string) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"tmpl": tmpl,
			"expr": expr,
		}).Info("running xml_xpath")
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
