package config

import (
	"github.com/go-redis/redis/v8"

	"github.com/creamsensation/cache/memory"
)

type Cache struct {
	Memory *memory.Client
	Redis  *redis.Client
}
