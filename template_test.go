package openmock

import (
	"net/http"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestContextMerge(t *testing.T) {
	t.Run("happy path example", func(t *testing.T) {
		orig_context := Context{
			HTTPBody: "hello",
			HTTPPath: "goodbye",
		}
		merge_context := Context{
			HTTPPath:   "au revoir",
			KafkaTopic: "this is a topic",
		}
		expected_result := Context{
			HTTPPath:   "au revoir",
			KafkaTopic: "this is a topic",
			HTTPBody:   "hello",
		}

		actual_result := orig_context.Merge(merge_context)
		assert.Equal(t, expected_result, actual_result)
	})
}

func TestTemplateRender(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		raw := `{
				"transaction_id": "{{.KafkaPayload | jsonPath "transaction_id"}}",
				"first_name": "{{.AMQPPayload | xmlPath "user/first_name"}}",
				"middle_name": "{{.HTTPBody | jsonPath "user/middle_name"}}",
				"user": {{.HTTPBody | gJsonPath "user"}}
				}`
		c := Context{
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
				"middle_name": "H",
				"user": {"middle_name": "H"}
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
		c := Context{}
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
		c := Context{
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
		c := Context{}
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
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Equal(t, r, `    ONE TWO `)
	})

	t.Run("predefined templates reused in other templates", func(t *testing.T) {
		t1 := `{{define "T1"}}T1{{end}}`
		t2 := `{{define "T2"}}T2{{end}}`
		t3 := `{{template "T1"}}{{template "T2"}}`

		c := Context{}
		var err error
		_, err = c.Render(t1)
		assert.NoError(t, err)
		_, err = c.Render(t2)
		assert.NoError(t, err)
		r, err := c.Render(t3)
		assert.NoError(t, err)
		assert.Equal(t, r, `T1T2`)
	})

	t.Run("predefined templates reused in other templates with context", func(t *testing.T) {
		t1 := `{{define "T1"}}{{.HTTPPath}}{{end}}`
		t2 := `{{define "T2"}} and T2{{end}}`
		t3 := `{{template "T1" .}}{{template "T2"}}`

		c := Context{
			HTTPPath: "/path_to_t1",
		}
		var err error
		_, err = c.Render(t1)
		assert.NoError(t, err)
		_, err = c.Render(t2)
		assert.NoError(t, err)
		r, err := c.Render(t3)
		assert.NoError(t, err)
		assert.Equal(t, r, `/path_to_t1 and T2`)
	})

	t.Run("templates reused in other templates", func(t *testing.T) {
		t1 := `{{define "T1"}}{{.HTTPPath}}{{end}}`
		t2 := `{{define "T2"}}{{template "T1" .}}{{end}}` // use T1 in definition of T2
		t3 := `{{template "T2" .}}`

		c := Context{
			HTTPPath: "/path_to_t1",
		}
		var err error
		_, err = c.Render(t1)
		assert.NoError(t, err)
		_, err = c.Render(t2)
		assert.NoError(t, err)
		r, err := c.Render(t3)
		assert.NoError(t, err)
		assert.Equal(t, r, `/path_to_t1`)
	})

	t.Run("values used in body", func(t *testing.T) {
		template := `{{.Values.foo}}`
		values := map[string]interface{}{"foo": "bar"}

		context := Context{
			HTTPPath: "/path_to_t1",
			Values:   values,
		}
		rendered, _ := context.Render(template)
		assert.Equal(t, rendered, `bar`)
	})

	t.Run("missing values cause an error", func(t *testing.T) {
		template := `{{.Values.banana}}`
		values := map[string]interface{}{}

		context := Context{
			HTTPPath: "/path_to_t1",
			Values:   values,
		}
		_, err := context.Render(template)
		assert.Error(t, err)
	})
}

func TestRenderConditions(t *testing.T) {
	t.Run("match conditions happy code path", func(t *testing.T) {
		raw := `{{.AMQPPayload | jsonPath "dl_number" | contains "K1111111" | not}}`
		c := Context{
			AMQPPayload: `{"dl_number": "K3333333"}`,
		}
		assert.True(t, c.MatchCondition(raw))
	})

	t.Run("match conditions not eq", func(t *testing.T) {
		raw := `{{eq (.AMQPPayload | jsonPath "dl_number") "12345678"}}`
		c := Context{
			AMQPPayload: `{"dl_number": "K1111111"}`,
		}
		assert.False(t, c.MatchCondition(raw))
	})

	t.Run("match conditions when condition is empty string", func(t *testing.T) {
		raw := ""
		c := Context{}
		assert.True(t, c.MatchCondition(raw))
	})
}

func TestRenderRedis(t *testing.T) {
	t.Run("Set redis in template", func(t *testing.T) {
		raw := `{{.HTTPHeader.Get "X-TOKEN" | redisDo "SET" "k1"}}`
		c := Context{
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
		c := Context{om: &OpenMock{}}

		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Contains(t, r, "-4")
	})
}
