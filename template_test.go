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

	t.Run("temp templates defined in templates", func(t *testing.T) {
		raw := `
{{define "T1"}}ONE{{end}}
{{define "T2"}}TWO{{end}}
{{define "T3"}}{{template "T1"}} {{template "T2"}}{{end}}
{{template "T3"}}
`
		c := &Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Equal(t, r, `    ONE TWO `)
	})

	t.Run("predefined templates reused in other templates", func(t *testing.T) {
		t1 := `{{define "T1"}}T1{{end}}`
		t2 := `{{define "T2"}}T2{{end}}`
		t3 := `{{template "T1"}}{{template "T2"}}`

		c := &Context{}
		var err error
		_, err = c.Render(t1)
		assert.NoError(t, err)
		_, err = c.Render(t2)
		assert.NoError(t, err)
		r, err := c.Render(t3)
		assert.NoError(t, err)
		assert.Equal(t, r, `T1T2`)
	})
}

func TestRenderCondition(t *testing.T) {
	t.Run("match condition happy code path", func(t *testing.T) {
		raw := `{{eq (.AMQPPayload | jsonPath "dl_number") "K2879030"}}`
		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030"}`,
		}
		assert.True(t, c.MatchCondition(raw))
	})

	t.Run("match condition not eq", func(t *testing.T) {
		raw := `{{eq (.AMQPPayload | jsonPath "dl_number") "12345678"}}`
		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030"}`,
		}
		assert.False(t, c.MatchCondition(raw))
	})

	t.Run("match condition when condition is empty string", func(t *testing.T) {
		raw := ""
		c := &Context{}
		assert.True(t, c.MatchCondition(raw))
	})
}

func TestRenderConditions(t *testing.T) {
	t.Run("match conditions happy code path with multiple conditions", func(t *testing.T) {
		conditions := [2]string{
			`{{ eq (.AMQPPayload | jsonPath "dl_number") "K2879030" }}`,
			`{{ eq (.AMQPPayload | jsonPath "state") "CA" }}`,
		}

		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030", "state": "CA" }`,
		}
		assert.True(t, c.MatchConditions(conditions[:]))
	})

	t.Run("match conditions happy code path with single condition", func(t *testing.T) {
		conditions := [1]string{
			`{{ eq (.AMQPPayload | jsonPath "dl_number") "K2879030" }}`,
		}

		c := &Context{
			AMQPPayload: `{ "dl_number": "K2879030" }`,
		}
		assert.True(t, c.MatchConditions(conditions[:]))
	})

	t.Run("match conditions with failing conditions", func(t *testing.T) {
		conditions := [2]string{
			`{{ eq (.AMQPPayload | jsonPath "dl_number") "A1234567" }}`,
			`{{ eq (.AMQPPayload | jsonPath "state") "AZ" }}`,
		}

		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030", "state": "CA" }`,
		}
		assert.False(t, c.MatchConditions(conditions[:]))
	})

	t.Run("match conditions with one failing and one passing condition", func(t *testing.T) {
		conditions := [2]string{
			`{{ eq (.AMQPPayload | jsonPath "dl_number") "K2879030" }}`,
			`{{ eq (.AMQPPayload | jsonPath "state") "NY" }}`,
		}

		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030", "state": "CA" }`,
		}
		assert.False(t, c.MatchConditions(conditions[:]))
	})

	t.Run("match conditions with no conditions", func(t *testing.T) {
		conditions := [0]string{}
		c := &Context{}
		assert.True(t, c.MatchConditions(conditions[:]))
	})

	t.Run("match conditions when conditions are empty strings", func(t *testing.T) {
		conditions := [2]string{"", ""}
		c := &Context{}
		assert.True(t, c.MatchConditions(conditions[:]))
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
