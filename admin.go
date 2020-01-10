package openmock

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"syscall"
	"time"

	em "github.com/dafiti/echo-middleware"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/teamwork/reload"
	"gopkg.in/yaml.v2"
)

const reloadDelay = time.Second

func getRedisKey(setKey string) (redisKey string) {
	redisKey = redisTemplatesStore

	if setKey != "" {
		redisKey = redisKey + "_" + setKey
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

		setKey := c.Param("set_key")
		err = AddTemplates(om, mocks, setKey, shouldRestart)
		if err != nil {
			return err
		}

		return c.String(200, string(b))
	}
}

func AddTemplates(om *OpenMock, mocks []*Mock, setKey string, shouldRestart bool) error {
	redisKey := getRedisKey(setKey)

	for _, mock := range mocks {
		s, _ := yaml.Marshal([]*Mock{mock})
		_, err := om.redis.Do("HSET", redisKey, mock.Key, s)
		if err != nil {
			return err
		}
	}

	if shouldRestart {
		ReloadModel(om)
	}
	return nil
}

func DeleteTemplates(om *OpenMock, shouldRestart bool) func(c echo.Context) error {
	return func(c echo.Context) error {
		redisKey := getRedisKey(c.Param("set_key"))

		_, err := om.redis.Do("DEL", redisKey)
		if err != nil {
			return err
		}

		if shouldRestart {
			ReloadModel(om)
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
			ReloadModel(om)
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

func ReloadModel(om *OpenMock) {
	time.AfterFunc(reloadDelay, func() {
		if om.TemplatesDirHotReload {
			reload.Exec()
		} else {

			executableLoc, err := executableLocation()
			if err != nil {
				panic(fmt.Sprintf("Cannot restart can't find executableLoc: %v", err))
			}

			err = reloadProcess(executableLoc)
			if err != nil {
				panic(fmt.Sprintf("Cannot restart: %v", err))
			}
		}
	})
}

// from reload; library doesn't allow you to use its process reload stuff
// without watching SOME files.  So just copy/paste that functionality here.
// https://github.com/Teamwork/reload/blob/master/reload.go

// reloadProcess replaces the current process with a new copy of itself.
func reloadProcess(binSelf string) error {
	return syscall.Exec(binSelf, append([]string{binSelf}, os.Args[1:]...), os.Environ())
}

// Get location to executable.
func executableLocation() (string, error) {
	bin := os.Args[0]
	if !filepath.IsAbs(bin) {
		var err error
		bin, err = os.Executable()
		if err != nil {
			return "", errors.Wrapf(err,
				"cannot get path to binary %q (launch with absolute path): %v",
				os.Args[0], err)
		}
	}
	return bin, nil
}
