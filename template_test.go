package openmock

import (
	"net/http"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestTemplateRender(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		raw := `{
				"transaction_id": "{{.KafkaPayload | jsonPath "transaction_id"}}",
				"first_name": "{{.AMQPPayload | xmlPath "user/first_name"}}",
				"middle_name": "{{.HTTPBody | jsonPath "user/middle_name"}}"
				}`
		c := &Context{
			HTTPBody:     `{"user": {"middle_name": "H"}}`,
			KafkaPayload: `{"transaction_id": "t1234"}`,
			AMQPPayload:  `<user><first_name>John</first_name></user>`,
		}

		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.JSONEq(t, r, `
			{
				"transaction_id": "t1234",
				"first_name": "John",
				"middle_name": "H"
			}
		`)
	})

	t.Run("nil http body", func(t *testing.T) {
		raw := `
			{
				"transaction_id": "{{.KafkaPayload | jsonPath "transaction_id"}}",
				"first_name": "{{.AMQPPayload | xmlPath "user/first_name"}}",
				"middle_name": "{{.HTTPBody | jsonPath "user/middle_name"}}"
			}
			`
		c := &Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.JSONEq(t, r, `
			{
				"transaction_id": "",
				"first_name": "",
				"middle_name": ""
			}
		`)
	})

	t.Run("no template variables", func(t *testing.T) {
		e := echo.New()
		raw := `
			{
				"transaction_id": "t1234",
				"name": "abcd"
			}
			`
		c := &Context{
			HTTPContext: e.NewContext(&http.Request{}, nil),
		}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.JSONEq(t, r, `
			{
				"transaction_id": "t1234",
				"name": "abcd"
			}
		`)
	})

	t.Run("xml in json", func(t *testing.T) {
		raw := `
{
	"payload": "
		<?xml version=\"1.0\" encoding=\"UTF-8\"?>
			<Record>
			<r1>
			test
			</r1>
		</Record>
	"
}
			`
		c := &Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.JSONEq(t, r, `
			{
				"payload": "   <?xml version=\"1.0\" encoding=\"UTF-8\"?>    <Record>    <r1>    test    </r1>   </Record>  "
			}
		`)
	})
}

func TestRenderConditions(t *testing.T) {
	t.Run("match conditions happy code path", func(t *testing.T) {
		raw := `{{eq (.AMQPPayload | jsonPath "dl_number") "K2879030"}}`
		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030"}`,
		}
		assert.True(t, c.MatchCondition(raw))
	})

	t.Run("match conditions not eq", func(t *testing.T) {
		raw := `{{eq (.AMQPPayload | jsonPath "dl_number") "12345678"}}`
		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030"}`,
		}
		assert.False(t, c.MatchCondition(raw))
	})

	t.Run("match conditions when condition is empty string", func(t *testing.T) {
		raw := ""
		c := &Context{}
		assert.True(t, c.MatchCondition(raw))
	})
}

func TestRenderRedis(t *testing.T) {
	t.Run("Set redis in template", func(t *testing.T) {
		raw := `{{.HTTPHeader.Get "X-TOKEN" | redisDo "SET" "k1"}}`
		c := &Context{
			om:         &OpenMock{},
			HTTPHeader: http.Header{},
		}
		c.HTTPHeader.Set("X-TOKEN", "t123")
		_, err := c.Render(raw)
		assert.NoError(t, err)
		v, _ := redis.String(c.om.redis.Do("GET", "k1"))
		assert.Equal(t, "t123", v)
	})

	t.Run("Render arrays in template", func(t *testing.T) {
		raw := `
          {
			"setup1": "{{redisDo "RPUSH" "random" uuidv4}}",
			"setup2": "{{redisDo "RPUSH" "random" uuidv4}}",
			"setup3": "{{redisDo "RPUSH" "random" uuidv4}}",
            "random": [
              {{range $i, $v := redisDo "LRANGE" "random" 0 -1 | split ";;"}}
                "{{$v}}"
              {{end}}
            ]
          }
		  `
		c := &Context{om: &OpenMock{}}

		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Contains(t, r, "-4")
	})
}
