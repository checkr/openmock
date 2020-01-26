package openmock

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

const redisStringsSeparator = ";;"

type (
	// RedisDoer can run redis commands
	RedisDoer interface {
		Do(commandName string, args ...interface{}) (reply interface{}, err error)
	}

	redisDoer struct {
		pool *redis.Pool
	}
)

// NewRedisDoer creates a new RedisDoer
func NewRedisDoer(redisURL string) RedisDoer {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.DialURL(redisURL)
			if err != nil {
				logrus.Errorf("cannot connect to redis %s. err: %v", redisURL, err)
				return nil, err
			}
			return conn, nil
		},
	}
	return &redisDoer{pool: pool}
}

func (rd *redisDoer) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	conn := rd.pool.Get()
	defer conn.Close()
	return conn.Do(commandName, args...)
}

// SetRedis sets the Redis store for OpenMock
func (om *OpenMock) SetRedis() {
	switch om.RedisType {
	case "redis":
		om.redis = NewRedisDoer(om.RedisURL)
	default:
		s, err := miniredis.Run()
		if err != nil {
			logrus.Fatalf("cannot create miniredis. url:%s, err:%s", om.RedisURL, err)
		}
		om.redis = NewRedisDoer(fmt.Sprintf("redis://%s:%s", s.Host(), s.Port()))
	}
}

func (om *OpenMock) RedisDo(commandName string, args ...interface{}) (reply interface{}, err error) {
	if om.redis == nil {
		return "", errors.New("RedisDo before redis set up")
	}
	return om.redis.Do(commandName, args...)
}

func redisHandleReply(r interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	strs := cast.ToStringSlice(r)
	return strings.Join(strs, redisStringsSeparator), nil
}

func redisDo(om *OpenMock) func(keyAndArgs ...interface{}) interface{} {
	bannedRedisKeysRegex, err := regexp.Compile(redisTemplatesStore + `(_[\w+])?`)

	if err != nil || om == nil {
		return func(keyAndArgs ...interface{}) interface{} {
			return nil
		}
	}
	if om.redis == nil {
		om.SetRedis()
	}
	return func(keyAndArgs ...interface{}) interface{} {
		name, args := keyAndArgs[0], keyAndArgs[1:]

		for _, arg := range args {
			if v, ok := arg.(string); ok {
				if bannedRedisKeysRegex.MatchString(v) {
					return fmt.Errorf("may not redisDo operate on reserved keys (matching regex (%s))", bannedRedisKeysRegex.String())
				}
			}
		}

		v, err := redisHandleReply(om.redis.Do(name.(string), args...))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"cmd": keyAndArgs,
				"err": err,
			}).Errorf("failed to run redisDo")
		}
		return v
	}
}
