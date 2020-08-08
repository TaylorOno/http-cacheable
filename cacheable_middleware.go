package cacheable

import (
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

type HTTPCacheProvider interface {
	Get(string) (http.Response, bool)
	Set(string, http.Response, time.Duration)
}

const (
	DefaultExpiration time.Duration = 0
)

// NewCacheableMiddleware - Creates middleware that can be used to create cache enabled HTTP clients
func NewCacheableMiddleware(c HTTPCacheProvider) Middleware {
	return func(client Client) Client {
		return ClientFunc(func(req *http.Request) (*http.Response, error) {
			var response *http.Response

			cacheResult, ok := c.Get("test")
			if ok {
				return &cacheResult, nil
			}

			response, err := client.Do(req)
			if err != nil {
				return response, err
			}

			c.Set("key", *response, DefaultExpiration)

			return response, nil
		})
	}
}
