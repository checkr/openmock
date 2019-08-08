package openmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONPath(t *testing.T) {
	var ret string
	var err error
	var tmpl string

	tmpl = `{"transaction_id": "t1234"}`
	ret, err = jsonPath("/transaction_id", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, "t1234", ret)

	tmpl = `{"transaction_id": "t1234"}`
	ret, err = jsonPath("/transaction_id/abc", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, "", ret)

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = jsonPath("/user/first_name", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, "John", ret)

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = jsonPath("/*/first_name", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, "John", ret)

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = jsonPath("//first_name", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, "John", ret)

	tmpl = `[{"jsonrpc":"2.0","method":"classify","params":["GUILTY"],"id":112879785776}]`
	ret, err = jsonPath("*[1]/method", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, "classify", ret)

	tmpl = `[{"jsonrpc":"2.0","method":"classify","params":["GUILTY"],"id":112879785776}]`
	ret, err = jsonPath("*[1]/id", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, "112879785776", ret)
}

func TestHelpers(t *testing.T) {
	t.Run("uuid4 helpers", func(t *testing.T) {
		raw := `{{ uuidv4 }}`
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Contains(t, r, "-4")
	})

	t.Run("uuid5 helpers", func(t *testing.T) {
		raw := `{{ "1234" | uuidv5 }}`
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Contains(t, r, "-5")
	})

	t.Run("isLastIndex helpers", func(t *testing.T) {
		raw := `{{ "abc;;def" | splitList ";;" | isLastIndex 1 }}`
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Contains(t, r, "true")
	})

	t.Run("regexFind helpers", func(t *testing.T) {
		raw := `{{ regexFind "foo.?" "seafood fool" }}`
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Equal(t, "food", r)
	})

	t.Run("regexFindFirstSubmatch helpers", func(t *testing.T) {
		t.Run("normal [a-z]", func(t *testing.T) {
			raw := `{{ "peach" | regexFindFirstSubmatch "p([a-z]+)ch" }}`
			c := Context{}
			r, err := c.Render(raw)
			assert.NoError(t, err)
			assert.Equal(t, "ea", r)
		})
		t.Run("\\w", func(t *testing.T) {
			raw := `{{ "peach" | regexFindFirstSubmatch "p(\\w+)ch" }}`
			c := Context{}
			r, err := c.Render(raw)
			assert.NoError(t, err)
			assert.Equal(t, "ea", r)
		})
	})

	t.Run("regexFindAllSubmatch helpers", func(t *testing.T) {
		raw := `{{ "peach" | regexFindAllSubmatch "p([a-z]+)ch" }}`
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Equal(t, "[peach ea]", r)
	})

	t.Run("regexFindAllSubmatch helpers with index", func(t *testing.T) {
		raw := `h{{index (.HTTPBody | regexFindAllSubmatch "(p)([a-z]+)ch") 2}}l`
		c := Context{
			HTTPBody: "peach",
		}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Equal(t, "heal", r)
	})

	t.Run("htmlEscapeString helpers", func(t *testing.T) {
		raw := `{{ "<note>
					<to>Tove</to>
					<from>Jani</from>
					<heading name=\"heading\">Reminder</heading>
					<body>Don't forget me this weekend!</body>
					</note>" | htmlEscapeString }}`
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.NotContains(t, r, "<")
		assert.NotContains(t, r, ">")
		assert.NotContains(t, r, "\"")
	})

	t.Run("hmacSHA256 helpers", func(t *testing.T) {
		raw := `{{ "some data to be signed" | hmacSHA256 "secret-key" }}`
		c := Context{}
		r, err := c.Render(raw)
		assert.NoError(t, err)
		assert.Equal(t, "a5e93ff4bc1ef4e57ade71c759617fec0897b7d480add9a13ae37694ce319aed", r)
	})
}
