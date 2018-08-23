package openmock

import (
	"fmt"
	"io/ioutil"

	em "github.com/dafiti/echo-middleware"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

func (om *OpenMock) startHTTP() {
	e := echo.New()
	e.HideBanner = true
	e.Use(em.Logrus())
	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		logrus.WithFields(logrus.Fields{
			"req": string(reqBody),
			"res": string(resBody),
		}).Info()
	}))

	mocks := om.repo.HTTPMocks
	for h, ms := range mocks {
		func(h ExpectHTTP, ms MocksArray) {
			e.Match(
				[]string{h.Method},
				h.Path,
				func(ec echo.Context) error {
					body, _ := ioutil.ReadAll(ec.Request().Body)
					c := &Context{
						HTTPContext:     ec,
						HTTPHeader:      ec.Request().Header,
						HTTPBody:        string(body),
						HTTPPath:        ec.Path(),
						HTTPQueryString: ec.QueryString(),
						om:              om,
					}
					return ms.DoActions(c)
				},
			)
		}(h, ms)
	}

	e.Logger.Fatal(
		e.Start(fmt.Sprintf("%s:%d", om.HTTPHost, om.HTTPPort)),
	)
}
