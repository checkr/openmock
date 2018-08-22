package openmock

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestTemplateRender(t *testing.T) {
	t.Run("happy code path", func(t *testing.T) {
		raw := `{
				"transaction_id": "{{json_xpath .KafkaPayload "transaction_id"}}",
				"first_name": "{{xml_xpath .AMQPPayload "user/first_name"}}",
				"middle_name": "{{json_xpath .HTTPBody "user/middle_name"}}"
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
				"transaction_id": "{{json_xpath .KafkaPayload "transaction_id"}}",
				"first_name": "{{xml_xpath .AMQPPayload "user/first_name"}}",
				"middle_name": "{{json_xpath .HTTPBody "user/middle_name"}}"
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
"payload": "<Record>
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
				"payload": "<Record><r1>test</r1></Record>"
			}
		`)
	})
}

func TestRenderConditions(t *testing.T) {
	t.Run("match conditions happy code path", func(t *testing.T) {
		raw := `{{eq (json_xpath .AMQPPayload "dl_number") "K2879030"}}`
		c := &Context{
			AMQPPayload: `{"dl_number": "K2879030"}`,
		}
		assert.True(t, c.MatchCondition(raw))
	})

	t.Run("match conditions not eq", func(t *testing.T) {
		raw := `{{eq (json_xpath .AMQPPayload "dl_number") "12345678"}}`
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

func TestJSONXPath(t *testing.T) {
	var ret string
	var err error
	var tmpl string

	tmpl = `{"transaction_id": "t1234"}`
	ret, err = JSONXPath(tmpl, "/transaction_id")
	assert.NoError(t, err)
	assert.Equal(t, ret, "t1234")

	tmpl = `{"transaction_id": "t1234"}`
	ret, err = JSONXPath(tmpl, "/transaction_id/abc")
	assert.NoError(t, err)
	assert.Equal(t, ret, "")

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = JSONXPath(tmpl, "/user/first_name")
	assert.NoError(t, err)
	assert.Equal(t, ret, "John")

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = JSONXPath(tmpl, "/*/first_name")
	assert.NoError(t, err)
	assert.Equal(t, ret, "John")

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = JSONXPath(tmpl, "//first_name")
	assert.NoError(t, err)
	assert.Equal(t, ret, "John")

	tmpl = `[{"jsonrpc":"2.0","method":"classify","params":["GUILTY"],"id":112879785776}]`
	ret, err = JSONXPath(tmpl, "*[1]/method")
	assert.NoError(t, err)
	assert.Equal(t, ret, "classify")

	tmpl = `[{"jsonrpc":"2.0","method":"classify","params":["GUILTY"],"id":112879785776}]`
	ret, err = JSONXPath(tmpl, "*[1]/id")
	assert.NoError(t, err)
	assert.Equal(t, ret, "112879785776")
}

func TestHelpers(t *testing.T) {
	t.Run("uuid4 helpers", func(t *testing.T) {
		raw := `{{ uuidv4 }}`
		c := &Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Contains(t, r, "-4")
	})

	t.Run("uuid4 helpers", func(t *testing.T) {
		raw := `{{ "1234" | uuidv5 }}`
		c := &Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Contains(t, r, "-5")
	})

	t.Run("regexFind helpers", func(t *testing.T) {
		raw := `{{ regexFind "foo.?" "seafood fool" }}`
		c := &Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Equal(t, r, "food")
	})
}
