package evaluator

import (
	"net/http"
	"net/http/httptest"
	"strings"

	om "github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
	"github.com/fatih/structs"
	"github.com/labstack/echo/v4"
)

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
