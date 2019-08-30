package openmock

import (
	"bytes"
	"fmt"
	"time"

	em "github.com/dafiti/echo-middleware"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/teamwork/reload"
	yaml "gopkg.in/yaml.v2"
)

const reloadDelay = time.Second

// StartAdmin starts an admin HTTP server
// that can CRUD the templates
func (om *OpenMock) StartAdmin() {
	if !om.AdminHTTPEnabled {
		return
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(em.Logrus())

	e.POST("/api/v1/templates", func(c echo.Context) error {
		body := c.Request().Body
		defer body.Close()

		buf := &bytes.Buffer{}

		_, err := buf.ReadFrom(body)
		if err != nil {
			return err
		}

		b := buf.Bytes()

		mocks := []*Mock{}
		if err := yaml.UnmarshalStrict(b, &mocks); err != nil {
			return c.String(400, fmt.Sprintf("not valid YAML %s", err))
		}

		for _, mock := range mocks {
			s, _ := yaml.Marshal([]*Mock{mock})
			_, err := om.redis.Do("HSET", redisTemplatesStore, mock.Key, s)
			if err != nil {
				return err
			}
		}

		time.AfterFunc(reloadDelay, func() { reload.Exec() })
		return c.String(200, string(b))
	})

	e.GET("/api/v1/templates", func(c echo.Context) error {
		return c.String(200, string(om.repo.ToYAML()))
	})

	go func() {
		logrus.Fatal(
			e.Start(fmt.Sprintf(
				"%s:%d",
				om.AdminHTTPHost,
				om.AdminHTTPPort,
			)),
		)
	}()
}
