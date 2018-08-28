package openmock

import (
	"strings"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

const redisStringsSeparator = ";;"

func parseCommand(cmd string) (name string, args []interface{}) {
	cmds := strings.Split(cmd, " ")
	if len(cmds) == 1 {
		return cmds[0], nil
	}
	for _, a := range cmds[1:] {
		args = append(args, a)
	}
	return cmds[0], args
}

// RedisDoer can run redis commands
type RedisDoer interface {
	Do(commandName string, args ...interface{}) (reply interface{}, err error)
}

// SetRedis sets the Redis store for OpenMock
func (om *OpenMock) SetRedis() {
	switch om.RedisType {
	case "redis":
		client, err := redis.Dial("tcp", om.RedisURL)
		if err != nil {
			logrus.Fatalf("cannot connect to redis %s. err: %v", om.RedisURL, err)
		}
		om.redis = client
	default:
		s, err := miniredis.Run()
		if err != nil {
			logrus.Fatalf("cannot create miniredis", om.RedisURL, err)
		}
		client, err := redis.Dial("tcp", s.Addr())
		if err != nil {
			logrus.Fatalf("cannot connect to miniredis %s. err: %v", s.Addr(), err)
		}
		om.redis = client
	}
}

func redisHandleReply(r interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	strs := cast.ToStringSlice(r)
	return strings.Join(strs, redisStringsSeparator), nil
}

func redisDo(om *OpenMock) func(keyAndArgs ...interface{}) interface{} {
	if om == nil {
		return func(keyAndArgs ...interface{}) interface{} {
			return nil
		}
	}
	if om.redis == nil {
		om.SetRedis()
	}
	return func(keyAndArgs ...interface{}) interface{} {
		name, args := keyAndArgs[0], keyAndArgs[1:]
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
