package openmock

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func genLocalFuncMap(om *OpenMock) template.FuncMap {
	return template.FuncMap{
		"jsonPath":    jsonPath,
		"xmlPath":     xmlPath,
		"uuidv5":      uuidv5,
		"redisDo":     redisDo(om),
		"isLastIndex": isLastIndex,
	}
}

func jsonPath(expr string, tmpl string) (ret string, err error) {
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

func xmlPath(expr string, tmpl string) (ret string, err error) {
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

func isLastIndex(i interface{}, a interface{}) bool {
	spew.Dump(a)
	return i == reflect.ValueOf(a).Len()-1
}
