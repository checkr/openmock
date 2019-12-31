package openmock

import (
	"bytes"
	"fmt"
	"net/url"
	"time"

	em "github.com/dafiti/echo-middleware"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/teamwork/reload"
	"gopkg.in/yaml.v2"
)

const reloadDelay = time.Second

func getRedisKey(c echo.Context) (redisKey string) {
	redisKey = redisTemplatesStore
	alternativeKey := c.Param("set_key")
	if alternativeKey != "" {
		redisKey = redisKey + "_" + alternativeKey
	}
	return redisKey
}

func PostTemplates(om *OpenMock, shouldRestart bool) func(c echo.Context) error {
	return func(c echo.Context) error {
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

		redisKey := getRedisKey(c)

		for _, mock := range mocks {
			s, _ := yaml.Marshal([]*Mock{mock})
			_, err := om.redis.Do("HSET", redisKey, mock.Key, s)
			if err != nil {
				return err
			}
		}

		if shouldRestart {
			time.AfterFunc(reloadDelay, func() { reload.Exec() })
		}
		return c.String(200, string(b))
	}
}

func DeleteTemplates(om *OpenMock, shouldRestart bool) func(c echo.Context) error {
	return func(c echo.Context) error {
		redisKey := getRedisKey(c)

		_, err := om.redis.Do("DEL", redisKey)
		if err != nil {
			return err
		}

		if shouldRestart {
			time.AfterFunc(reloadDelay, func() { reload.Exec() })
		}
		return c.String(204, "")
	}
}

func DeleteTemplateByKey(om *OpenMock, shouldRestart bool) func(c echo.Context) error {
	return func(c echo.Context) error {
		key, err := url.QueryUnescape(c.Param("key"))
		if err != nil {
			return c.String(400, fmt.Sprintf("invalid key: %v. error: %s", key, err))
		}

		v, err := om.redis.Do("HGET", redisTemplatesStore, key)
		m, err := redis.Bytes(v, err)

		if err != nil {
			return err
		}

		if m == nil {
			return c.String(404, fmt.Sprintf("key not found: %v", key))
		}

		_, err = om.redis.Do("HDEL", redisTemplatesStore, key)
		if err != nil {
			return err
		}

		if shouldRestart {
			time.AfterFunc(reloadDelay, func() { reload.Exec() })
		}
		return c.String(200, fmt.Sprintf("deleted:\n\n%v", string(m)))
	}
}

func GetTemplates(om *OpenMock) func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.String(200, string(om.repo.ToYAML()))
	}
}

// StartAdmin starts an admin HTTP server
// that can CRUD the templates
func (om *OpenMock) StartAdmin() {
	if !om.AdminHTTPEnabled {
		return
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(em.Logrus())

	e.POST("/api/v1/templates", PostTemplates(om, true))
	e.POST("/api/v1/template_sets/:set_key", PostTemplates(om, true))
	e.DELETE("/api/v1/template_sets/:set_key", DeleteTemplates(om, true))
	e.DELETE("/api/v1/templates", DeleteTemplates(om, true))
	e.DELETE("/api/v1/templates/:key", DeleteTemplateByKey(om, true))
	e.GET("/api/v1/templates", GetTemplates(om))

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
