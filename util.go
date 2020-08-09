package cacheable

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// GenerateKeyHash - a generic method for creating a string cache key from an http.Request object
func GenerateKeyHash(r *http.Request) string {
	s := sha1.New()
	s.Write([]byte(r.Method))
	s.Write([]byte(r.URL.Host))
	s.Write([]byte(r.URL.Path))

	paramList := make([]string, 0, len(r.URL.Query()))
	for key, value := range r.URL.Query() {
		paramList = append(paramList, key+"="+strings.Join(value, ","))
	}
	if !sort.StringsAreSorted(paramList) {
		sort.Strings(paramList)
	}
	for _, value := range paramList {
		s.Write([]byte(value))
	}

	headersList := make([]string, 0, len(r.Header))
	for key, value := range r.Header {
		headersList = append(headersList, key+"="+strings.Join(value, ","))
	}
	if !sort.StringsAreSorted(headersList) {
		sort.Strings(headersList)
	}
	for _, value := range headersList {
		s.Write([]byte(value))
	}

	if r.Body != nil {
		requestBody, err := r.GetBody()
		if err != nil {
			defer requestBody.Close()
		}

		body, _ := ioutil.ReadAll(requestBody)
		if len(body) > 0 {
			s.Write(body)
		}
	}

	return base64.StdEncoding.EncodeToString(s.Sum(nil))
}

// ContextWithCacheConfig - Returns a context with a CacheConfig Object
func ContextWithCacheConfig(ctx context.Context, config CacheConfig) context.Context {
	return context.WithValue(ctx, cacheConfigKey, config)
}