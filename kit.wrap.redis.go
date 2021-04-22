package kitgo

import (
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

// Redis implement a wrapper around "go-redis/redis"
var Redis redis_

type redis_ struct{}

func (redis_) New(conf *RedisConfig) *RedisWrapper {
	return &RedisWrapper{redis.NewUniversalClient(conf)}
}

func (redis_) Test() (*RedisWrapper, *RedisMock) {
	db, mock := redismock.NewClientMock()
	return &RedisWrapper{db}, &RedisMock{mock}
}

type RedisConfig = redis.UniversalOptions
type RedisWrapper struct{ redis.UniversalClient }
type RedisMock struct{ redismock.ClientMock }
