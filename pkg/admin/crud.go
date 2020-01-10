package admin

import (
	"fmt"

	"github.com/checkr/openmock"
	"github.com/checkr/openmock/swagger_gen/models"
	"github.com/checkr/openmock/swagger_gen/restapi/operations/template"
	"github.com/checkr/openmock/swagger_gen/restapi/operations/template_set"
	"github.com/go-openapi/runtime/middleware"

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

	err := openmock.AddTemplates(c.om, omMocks, "", true)
	if err != nil {
		logrus.Errorf("Couldn't add OM templates %v", err)
		resp := template.NewPostTemplatesBadRequest()
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: &message})
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
		resp.SetPayload(&models.Error{Message: &message})
		return resp
	}

	openmock.ReloadModel(c.om)

	resp := template.NewDeleteTemplatesNoContent()
	return resp
}

func (c *crud) DeleteTemplate(params template.DeleteTemplateParams) middleware.Responder {
	v, err := c.om.RedisDo("HGET", c.om.RedisTemplatesStore(), params.TemplateKey)
	m, err := redis.Bytes(v, err)

	if err != nil {
		resp := template.NewDeleteTemplateDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: &message})
		return resp
	}

	if m == nil {
		resp := template.NewDeleteTemplateNotFound()
		message := fmt.Sprintf("key not found: %s", params.TemplateKey)
		resp.SetPayload(&models.Error{Message: &message})
		return resp
	}

	_, err = c.om.RedisDo("HDEL", c.om.RedisTemplatesStore(), params.TemplateKey)
	if err != nil {
		resp := template.NewDeleteTemplateDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: &message})
		return resp
	}

	openmock.ReloadModel(c.om)

	resp := template.NewDeleteTemplateNoContent()
	return resp
}

func (c *crud) PostTemplateSet(params template_set.PostTemplateSetParams) middleware.Responder {
	swaggerMocks := params.Mocks
	omMocks := swaggerToOpenmockMockArray(swaggerMocks)

	err := openmock.AddTemplates(c.om, omMocks, params.SetKey, true)
	if err != nil {
		logrus.Errorf("Couldn't add OM templates %v", err)
		resp := template_set.NewPostTemplateSetBadRequest()
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: &message})
		return resp
	}

	resp := template_set.NewPostTemplateSetOK()
	resp.SetPayload(swaggerMocks)
	return resp
}

func (c *crud) DeleteTemplateSet(params template_set.DeleteTemplateSetParams) middleware.Responder {
	setKey := fmt.Sprintf("%s_%s", c.om.RedisTemplatesStore(), params.SetKey)
	logrus.Infof("Deleting Template set %s", setKey)
	_, err := c.om.RedisDo("DEL", setKey)
	if err != nil {
		resp := template_set.NewDeleteTemplateSetDefault(500)
		message := fmt.Sprintf("%v", err)
		resp.SetPayload(&models.Error{Message: &message})
		return resp
	}

	openmock.ReloadModel(c.om)

	resp := template_set.NewDeleteTemplateSetNoContent()
	return resp
}
