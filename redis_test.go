package openmock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	om := &OpenMock{}
	om.SetRedis()

	t.Run("get non-exists data", func(t *testing.T) {
		v, err := redisHandleReply(om.redis.Do("get", "non-exist"))
		assert.NoError(t, err)
		assert.Empty(t, v)
	})

	t.Run("set data", func(t *testing.T) {
		v, err := redisHandleReply(om.redis.Do("set", "hello", "world"))
		assert.NoError(t, err)
		assert.NotEmpty(t, v)
	})

	t.Run("get data", func(t *testing.T) {
		v, err := redisHandleReply(om.redis.Do("get", "hello"))
		assert.NoError(t, err)
		assert.Equal(t, "world", v)
	})

	t.Run("rpush set array data", func(t *testing.T) {
		v, err := redisHandleReply(om.redis.Do("rpush", "k1", "v1"))
		v, err = redisHandleReply(om.redis.Do("rpush", "k1", "v2"))
		assert.NoError(t, err)
		assert.NotEmpty(t, v)
	})

	t.Run("rpush get array data", func(t *testing.T) {
		v, err := redisHandleReply(om.redis.Do("lrange", "k1", 0, -1))
		assert.NoError(t, err)
		assert.Equal(t, "v1;;v2", v)
	})
}
