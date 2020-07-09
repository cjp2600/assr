package cache

import (
	"github.com/cjp2600/assr/config"
	"github.com/patrickmn/go-cache"
)

var client *cache.Cache

func GetClient() *cache.Cache {
	if client == nil {
		client = cache.New(config.GetCacheDurationTTL(), config.GetCacheDurationTTL())
	}
	return client
}
