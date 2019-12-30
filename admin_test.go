package openmock

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func getTestOM(t *testing.T) *OpenMock {
	om := &OpenMock{
		TemplatesDir: "demo_templates",
	}
	om.SetupRepo()
	err := om.Load()
	assert.NoError(t, err)
	om.RedisType = ""
	om.SetRedis()
	return om
}

func testRequest(method, path string, e *echo.Echo) (int, string) {
	return testRequestBody(method, path, e, nil)
}

func testRequestBody(method, path string, e *echo.Echo, body io.Reader) (int, string) {
	return testRequestFull(method, path, e, body, map[string]string{})
}

func testRequestFull(method, path string, e *echo.Echo, body io.Reader, headers map[string]string) (int, string) {
	req := httptest.NewRequest(method, path, body)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestGetTemplates(t *testing.T) {
	t.Run("Get Templates returns YAML", func(t *testing.T) {
		om := getTestOM(t)
		handler := GetTemplates(om)

		assert.NotNil(t, handler)

		e := echo.New()
		e.GET("/", handler)
		c, b := testRequest(http.MethodGet, "/", e)
		assert.Equal(t, http.StatusOK, c)

		mocks := []*Mock{}
		if err := yaml.UnmarshalStrict([]byte(b), &mocks); err != nil {
			t.FailNow()
		}

		for _, m := range mocks {
			if m.Key == "ping" {
				assert.Equal(t, m, om.repo.Behaviors["ping"])
			}
		}
	})
}

func TestDeleteTemplateByKey(t *testing.T) {
	om := getTestOM(t)
	handler := DeleteTemplateByKey(om, false)
	e := echo.New()
	e.DELETE("/:key", handler)

	t.Run("Delete template at key deletes it", func(t *testing.T) {
		_, err := om.redis.Do("HSET", redisTemplatesStore, "123", "stuff")
		if err != nil {
			t.FailNow()
		}

		c, b := testRequest(http.MethodDelete, "/123", e)
		assert.Equal(t, http.StatusOK, c)
		assert.NotEmpty(t, b)

		v, err := om.redis.Do("HGET", redisTemplatesStore, "123")
		result, err := redis.Bytes(v, err)
		assert.NotEmpty(t, err)
		assert.Equal(t, string(result), "")
	})
	t.Run("Delete non-existing key", func(t *testing.T) {
		c, _ := testRequest(http.MethodDelete, "/123", e)
		// assert.Equal(t, http.StatusNotFound, c) // TODO catch exception in echo when running redigo
		assert.Equal(t, http.StatusInternalServerError, c)
	})
}

func TestPostTemplates(t *testing.T) {
	om := getTestOM(t)
	handler := PostTemplates(om, false)
	e := echo.New()
	e.POST("/", handler)

	bodyString := `
- key: 123
  kind: Behavior
  expect:
    http:
      method: GET
      path: /ping
  actions:
    - reply_http:
        status_code: 200
        body: OK
        headers:
          Content-Type: text/xml	
  `

	t.Run("Post Happy path", func(t *testing.T) {
		bodyReader := strings.NewReader(bodyString)
		c, b := testRequestBody(http.MethodPost, "/", e, bodyReader)
		assert.Equal(t, http.StatusOK, c)
		assert.NotEmpty(t, b)

		v, err := om.redis.Do("HGET", redisTemplatesStore, "123")
		result, err := redis.Bytes(v, err)
		assert.Empty(t, err)
		assert.NotEmpty(t, string(result))
	})

	t.Run("Post with alternate key header", func(t *testing.T) {
		postKey := "foobar"
		headers := map[string]string{
			postKeyHeader: postKey,
		}
		bodyReader := strings.NewReader(bodyString)
		c, b := testRequestFull(http.MethodPost, "/", e, bodyReader, headers)
		assert.Equal(t, http.StatusOK, c)
		assert.NotEmpty(t, b)

		v, err := om.redis.Do("HGET", redisTemplatesStore+"_"+postKey, "123")
		result, err := redis.Bytes(v, err)
		assert.Empty(t, err)
		assert.NotEmpty(t, string(result))
	})
}
