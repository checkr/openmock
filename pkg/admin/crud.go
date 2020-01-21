package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
	"github.com/checkr/openmock/swagger_gen/restapi/operations/template"
	"github.com/checkr/openmock/swagger_gen/restapi/operations/template_set"
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	"github.com/teamwork/reload"
	yaml "gopkg.in/yaml.v2"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

type CRUD interface {
	// Templates
	GetTemplates(template.GetTemplatesParams) middleware.Responder
	PostTemplates(template.PostTemplatesParams) middleware.Responder
	DeleteTemplates(template.DeleteTemplatesParams) middleware.Responder
	DeleteTemplate(template.DeleteTemplateParams) middleware.Responder

	// Template Sets
	PostTemplateSet(template_set.PostTemplateSetParams) middleware.Responder
	DeleteTemplateSet(template_set.DeleteTemplateSetParams) middleware.Responder
}

func NewCRUD(om *openmock.OpenMock) CRUD {
	return &crud{om: om}
}

type crud struct {
	om *openmock.OpenMock
}

func (c *crud) GetTemplates(params template.GetTemplatesParams) middleware.Responder {
	resp := template.NewGetTemplatesOK()
	allMocks := c.om.ToArray()

	swaggerMocks := openmockToSwaggerMockArray(allMocks)

	resp.SetPayload(swaggerMocks)
	return resp
}

func (c *crud) PostTemplates(params template.PostTemplatesParams) middleware.Responder {
	swaggerMocks := params.Mocks
	omMocks := swaggerToOpenmockMockArray(swaggerMocks)

	err := c.addTemplates(omMocks, "", true)
	if err != nil {
		logrus.Errorf("Couldn't add OM templates %v", err)
		resp := template.NewPostTemplatesBadRequest()
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}

	resp := template.NewPostTemplatesOK()
	resp.SetPayload(swaggerMocks)
	return resp
}

func (c *crud) DeleteTemplates(params template.DeleteTemplatesParams) middleware.Responder {
	_, err := c.om.RedisDo("DEL", c.om.RedisTemplatesStore())
	if err != nil {
		resp := template.NewDeleteTemplatesDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}

	c.reloadModel()

	resp := template.NewDeleteTemplatesNoContent()
	return resp
}

func (c *crud) DeleteTemplate(params template.DeleteTemplateParams) middleware.Responder {
	v, err := c.om.RedisDo("HGET", c.om.RedisTemplatesStore(), params.TemplateKey)
	if err != nil {
		resp := template.NewDeleteTemplateDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}

	m, err := redis.Bytes(v, err)
	if err != nil {
		resp := template.NewDeleteTemplateDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}
	if m == nil {
		resp := template.NewDeleteTemplateNotFound()
		message := fmt.Sprintf("key not found: %s", params.TemplateKey)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}

	_, err = c.om.RedisDo("HDEL", c.om.RedisTemplatesStore(), params.TemplateKey)
	if err != nil {
		resp := template.NewDeleteTemplateDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}

	c.reloadModel()

	resp := template.NewDeleteTemplateNoContent()
	return resp
}

func (c *crud) PostTemplateSet(params template_set.PostTemplateSetParams) middleware.Responder {
	swaggerMocks := params.Mocks
	omMocks := swaggerToOpenmockMockArray(swaggerMocks)

	err := c.addTemplates(omMocks, params.SetKey, true)
	if err != nil {
		logrus.Errorf("Couldn't add OM templates %v", err)
		resp := template_set.NewPostTemplateSetBadRequest()
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}

	resp := template_set.NewPostTemplateSetOK()
	resp.SetPayload(swaggerMocks)
	return resp
}

func (c *crud) DeleteTemplateSet(params template_set.DeleteTemplateSetParams) middleware.Responder {
	setKey := c.getRedisKey(params.SetKey)
	logrus.Infof("Deleting Template set %s", setKey)
	_, err := c.om.RedisDo("DEL", setKey)
	if err != nil {
		resp := template_set.NewDeleteTemplateSetDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: message})
		return resp
	}

	c.reloadModel()

	resp := template_set.NewDeleteTemplateSetNoContent()
	return resp
}

func (c *crud) getRedisKey(setKey string) (redisKey string) {
	if setKey != "" {
		return fmt.Sprintf("%s_%s", c.om.RedisTemplatesStore(), setKey)
	}
	return c.om.RedisTemplatesStore()
}

const reloadDelay = time.Second

func (c *crud) addTemplates(mocks []*openmock.Mock, setKey string, shouldRestart bool) error {
	redisKey := c.getRedisKey(setKey)

	for _, mock := range mocks {
		s, _ := yaml.Marshal([]*openmock.Mock{mock})
		_, err := c.om.RedisDo("HSET", redisKey, mock.Key, s)
		if err != nil {
			return err
		}
	}

	if shouldRestart {
		c.reloadModel()
	}
	return nil
}

func (c *crud) reloadModel() {
	time.AfterFunc(reloadDelay, func() {
		if c.om.TemplatesDirHotReload {
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
