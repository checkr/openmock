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
	assert.Equal(t, ret, "t1234")

	tmpl = `{"transaction_id": "t1234"}`
	ret, err = jsonPath("/transaction_id/abc", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, ret, "")

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = jsonPath("/user/first_name", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, ret, "John")

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = jsonPath("/*/first_name", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, ret, "John")

	tmpl = `{"user": {"first_name": "John"}}`
	ret, err = jsonPath("//first_name", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, ret, "John")

	tmpl = `[{"jsonrpc":"2.0","method":"classify","params":["GUILTY"],"id":112879785776}]`
	ret, err = jsonPath("*[1]/method", tmpl)
	assert.NoError(t, err)
	assert.Equal(t, ret, "classify")

	tmpl = `[{"jsonrpc":"2.0","method":"classify","params":["GUILTY"],"id":112879785776}]`
	ret, err = jsonPath("*[1]/id", tmpl)
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

func TestRedisDo(t *testing.T) {
	om := &OpenMock{}
	r := redisDo(om)

	t.Run("get non-exists", func(t *testing.T) {
		v := r("GET", "non-exists")
		assert.Empty(t, v)
	})

	t.Run("set and then get", func(t *testing.T) {
		r("SET", "hello", "456")
		v := r("GET", "hello")
		assert.Equal(t, "456", v)
	})
}
