package cacheable

import (
	"context"
	"net/http"
	"time"
)

type Middleware func(client Client) Client

//go:generate mockgen -destination=mocks/cachable.go -package=mocks -source=cacheable_middleware.go
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientFunc func(req *http.Request) (*http.Response, error)

func (c ClientFunc) Do(req *http.Request) (*http.Response, error) {
	return c(req)
}

type key int

const cacheConfigKey key = iota

type CacheConfig struct {
	Key string
	TTL *int
}

type HTTPCacheProvider interface {
	Get(string) (*http.Response, bool)
	Set(string, *http.Response, time.Duration)
}

// NewCacheableMiddleware - Creates Middleware that can be used to create cache enabled HTTP clients
func NewCacheableMiddleware(c HTTPCacheProvider, ttl int) Middleware {
	defaultExpiration := time.Duration(ttl) * time.Second
	return func(client Client) Client {
		return ClientFunc(func(req *http.Request) (*http.Response, error) {
			var response *http.Response
			key := getKey(req)

			cacheResult, ok := c.Get(key)
			if ok {
				return cacheResult, nil
			}

			response, err := client.Do(req)
			if err != nil {
				return response, err
			}

			ttl := getTTL(req, defaultExpiration)
			c.Set(key, response, ttl)

			return response, nil
		})
	}
}

func getConfigFromContext(ctx context.Context) CacheConfig {
	value := ctx.Value(cacheConfigKey)
	config, ok := value.(CacheConfig)
	if !ok {
		return CacheConfig{}
	}
	return config
}

func getKey(r *http.Request) string {
	config := getConfigFromContext(r.Context())
	if len(config.Key) > 0 {
		return config.Key
	}
	return GenerateKeyHash(r)
}

func getTTL(r *http.Request, defaultTTL time.Duration) time.Duration {
	config := getConfigFromContext(r.Context())
	if config.TTL != nil {
		return time.Duration(*config.TTL) * time.Second
	}
	return defaultTTL
}
