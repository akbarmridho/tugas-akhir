package test_containers

import (
	"testing"
)

func Test_RedisCluster(t *testing.T) {
	t.Run("Redis cluster ok", func(t *testing.T) {
		GetRedisCluster(t)
	})
}
