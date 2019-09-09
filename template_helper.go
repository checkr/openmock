package openmock

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func genLocalFuncMap(om *OpenMock) template.FuncMap {
	return template.FuncMap{
		"htmlEscapeString":       template.HTMLEscapeString,
		"isLastIndex":            isLastIndex,
		"jsonPath":               jsonPath,
		"redisDo":                redisDo(om),
		"regexFindAllSubmatch":   regexFindAllSubmatch,
		"regexFindFirstSubmatch": regexFindFirstSubmatch,
		"uuidv5":                 uuidv5,
		"xmlPath":                xmlPath,
		"hmacSHA256":             hmacSHA256,
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
		}).Debug("running json xpath")
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
		}).Debug("running xml xpath")
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

// uuidv5 uses SHA1 and NameSpaceOID to generate consistent uuid
func uuidv5(dat string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(dat)).String()
}

// isLastIndex checks if the index is the last of the slice
// For example:
//  {{ range $i, $v := $arr }}
//    {{if isLastIndex $i $arr}}
//      "{{$v}}"
//    {{else}}
//      "{{$v}}",
//    {{end}}
//  {{end}}
func isLastIndex(i int, a interface{}) bool {
	return i == reflect.ValueOf(a).Len()-1
}

// regexFindAllSubmatch returns all the matching groups
// [0] string matches the whole regex
// [1:] strings matches the n-th group
func regexFindAllSubmatch(regex string, s string) []string {
	r := regexp.MustCompile(regex)
	return r.FindStringSubmatch(s)
}

// regexFindFirstSubmatch returns the first matching group
func regexFindFirstSubmatch(regex string, s string) string {
	matches := regexFindAllSubmatch(regex, s)
	if len(matches) <= 1 {
		return ""
	}
	return matches[1]
}

// hmacSHA256 computes SHA256 HMAC of data using secret
func hmacSHA256(secret string, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(data))
	if err != nil {
		logrus.WithField("err", err).Error("failed to hmacSHA256")
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
