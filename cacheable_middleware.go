//go:generate mockgen -destination=mocks/cachable.go -package=mocks -source=cacheable_middleware.go

package cacheable

import (
	"context"
	"net/http"
	"time"
)

const cacheConfigKey key = iota

type key int

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPCacheProvider is a cache interface that is used to cache http responses
type HTTPCacheProvider interface {
	Get(string) (*http.Response, bool)
	Set(string, *http.Response, time.Duration)
}

type Middleware func(client Client) Client

// Validator user defined function that should return true if the response should be cached.  See StatusCodeValidator
type Validator func(*http.Response) bool

type ClientFunc func(req *http.Request) (*http.Response, error)

func (c ClientFunc) Do(req *http.Request) (*http.Response, error) {
	return c(req)
}

// CacheConfig is an optional configuration that can be passed via context to alter caching behavior on a per request basis
type CacheConfig struct {
	Key        string
	TTLSeconds int
}

// ContextWithCacheConfig Returns a context with a CacheConfig. The cache config can be used to override the default ttl
// or to provide a custom cache key.
func ContextWithCacheConfig(ctx context.Context, config CacheConfig) context.Context {
	return context.WithValue(ctx, cacheConfigKey, config)
}

// NewCacheableMiddleware Given a HTTPCacheProvider, TTL in seconds, and a validator function will created a Middleware
// that can be used to create cache enabled HTTP clients. TTL is not enforced by cacheable middleware and must be
// enforced by the HTTPCacheProvider.  Cacheable middleware will not store an HTTP response that does not return true
// when passed to the Validator Function.
func NewCacheableMiddleware(c HTTPCacheProvider, ttlSeconds int, isValid Validator) Middleware {
	defaultExpiration := time.Duration(ttlSeconds) * time.Second
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

			if isValid(response) {
				ttl := getTTL(req, defaultExpiration)
				c.Set(key, response, ttl)
			}

			return response, nil
		})
	}
}

// StatusCodeValidator a simple Validator function that will return true if the http.Response status code is in the
// success range.
func StatusCodeValidator(r *http.Response) bool {
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return false
	}
	return true
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
	if config.TTLSeconds > 0 {
		return time.Duration(config.TTLSeconds) * time.Second
	}
	return defaultTTL
}
