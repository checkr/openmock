package evaluator

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
	"github.com/fatih/structs"
	"github.com/labstack/echo/v4"
)

var performReplyHTTPAction = func(context *om.Context, mock *om.ActionReplyHTTP) (*models.ReplyHTTPActionPerformed, error) {
	// initial output struct
	out := &models.ReplyHTTPActionPerformed{
		ContentType: echo.MIMEApplicationJSON,
		StatusCode:  strconv.Itoa(mock.StatusCode),
	}
	output_headers := make(map[string]string)

	// override content type if a header was provided in the mock
	if ct, ok := mock.Headers[echo.HeaderContentType]; ok {
		out.ContentType = ct
	}

	// render body
	body, err := context.Render(mock.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Problem rendering body for replyHTTPAction: %v", err))
	}
	out.Body = body

	// render headers
	for k, v := range mock.Headers {
		rendered, err := context.Render(v)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Problem rendering header! %s %s %v", k, v, err))
		}
		output_headers[k] = rendered
	}

	// additional headers provided by OM
	body_len := fmt.Sprintf("%d", len(body))
	output_headers["Content-Length"] = body_len
	output_headers["Content-Type"] = out.ContentType

	out.Headers = output_headers
	return out, nil
}

var httpToOpenmockConditionContext = func(context *models.EvalHTTPContext) (*om.Context, error) {
	if context == nil {
		return nil, errors.New("missing input context")
	}

	headers := map[string][]string{}

	if context.Headers != nil {
		contextHeaders, ok := context.Headers.(map[string]interface{})

		if !ok {
			return nil, errors.New(fmt.Sprintf("can't parse context headers %T", context.Headers))
		}

		for k, v := range contextHeaders {
			v_string, ok2 := v.(string)
			if !ok2 {
				continue
			}

			newV := make([]string, 1)
			newV[0] = v_string
			headers[k] = newV
		}
	}

	return &om.Context{
		HTTPBody:        context.Body,
		HTTPPath:        context.Path,
		HTTPQueryString: context.QueryString,
		HTTPHeader:      headers,
	}, nil
}

var checkHTTPCondition = func(context *models.EvalHTTPContext, mock *om.Mock) bool {
	if context == nil || structs.IsZero(*context) || mock == nil || structs.IsZero(mock.Expect.HTTP) {
		return false
	}

	// check that methods match, we can save some time not doing echo test in that case
	methods_match := strings.ToLower(context.Method) == strings.ToLower(mock.Expect.HTTP.Method)
	if !methods_match {
		return false
	}

	// create a mini echo server with the path / method of the expect
	paths_match := false
	e := echo.New()

	e.Match(
		[]string{mock.Expect.HTTP.Method},
		mock.Expect.HTTP.Path,
		func(ec echo.Context) error {
			paths_match = true
			return nil
		},
	)

	// create a request and a HTTP recorder to test that echo server
	req, err := http.NewRequest(strings.ToUpper(context.Method), context.Path, strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return false
	}
	rec := httptest.NewRecorder()

	// run the HTTP request and see the result we get
	e.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	// if the HTTP request was handled successfully paths_match should've been set
	return paths_match
}
