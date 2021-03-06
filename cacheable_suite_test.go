package cacheable_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/TaylorOno/golandreporter"
	cacheable "github.com/TaylorOno/http-cacheable"
	"github.com/TaylorOno/http-cacheable/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCacheable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithCustomReporters(t, "Cacheable Suite", []Reporter{golandreporter.NewGolandReporter()})
}

var _ = Describe("cacheable_middleware", func() {
	var (
		mockCtrl   *gomock.Controller
		mockClient *mocks.MockClient
		mockCache  *mocks.MockHTTPCacheProvider
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockClient(mockCtrl)
		mockCache = mocks.NewMockHTTPCacheProvider(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("NewCacheableMiddleware", func() {
		It("Creates_a_Middleware_function", func() {
			result := cacheable.NewCacheableMiddleware(mockCache, 0, func(response *http.Response) bool { return true })

			Expect(reflect.TypeOf(result).Name()).To(Equal("Middleware"))
		})
	})

	Context("CacheableMiddleware", func() {
		It("Returns a ClientFunc", func() {
			client := &http.Client{}
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 0, func(response *http.Response) bool { return true })

			result := cacheableMiddleware(client)

			Expect(reflect.TypeOf(result).Name()).To(Equal("ClientFunc"))
		})
	})

	Context("Cache ClientFunc", func() {
		It("Calls Parent Method and caches result on success", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 0, func(response *http.Response) bool { return true })
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)
			resp := &http.Response{}

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(resp, nil)
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any())

			_, _ = result.Do(req)
		})

		It("Returns and error on parent error", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 0, func(response *http.Response) bool { return true })
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("error"))

			_, err := result.Do(req)

			Expect(err).To(HaveOccurred())
		})

		It("Returns and error on parent error and does not set cache", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 0, func(response *http.Response) bool { return true })
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("error"))
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			_, err := result.Do(req)

			Expect(err).To(HaveOccurred())
		})

		It("Returns result but does not set cache on invalid response", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 0, cacheable.StatusCodeValidator)
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)
			resp := &http.Response{StatusCode: 500}

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(resp, nil)
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			_, err := result.Do(req)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Does not call parent method on cache hit", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 0, func(response *http.Response) bool { return true })
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)
			resp := &http.Response{}

			mockCache.EXPECT().Get(gomock.Any()).Return(resp, true)

			_, _ = result.Do(req)
		})

		It("Uses key from context if available", func() {
			var key string
			var ttl time.Duration
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 1, func(response *http.Response) bool { return true })
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)
			req = req.WithContext(cacheable.ContextWithCacheConfig(context.Background(), cacheable.CacheConfig{Key: "key"}))
			resp := http.Response{}

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(&resp, nil)
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(k string, v interface{}, t time.Duration) {
				key = k
				ttl = t
			})

			_, _ = result.Do(req)

			Expect(key).To(Equal("key"))
			Expect(ttl).To(Equal(1 * time.Second))
		})

		It("Uses ttl from context if available", func() {
			var key string
			var ttl time.Duration
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache, 0, func(response *http.Response) bool { return true })
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)
			req = req.WithContext(cacheable.ContextWithCacheConfig(context.Background(), cacheable.CacheConfig{Key: "key", TTLSeconds: 5}))
			resp := http.Response{}

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(&resp, nil)
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(k string, v interface{}, t time.Duration) {
				key = k
				ttl = t
			})

			_, _ = result.Do(req)
			Expect(key).To(Equal("key"))
			Expect(ttl).To(Equal(5 * time.Second))
		})
	})
})
