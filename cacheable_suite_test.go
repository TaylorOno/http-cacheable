package cacheable_test

import (
	"cacheable"
	"cacheable/mocks"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCacheable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cacheable Suite")
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
		It("Creates a Middleware function", func() {
			result := cacheable.NewCacheableMiddleware(mockCache)

			Expect(reflect.TypeOf(result).Name()).To(Equal("Middleware"))
		})
	})

	Context("CacheableMiddleware", func() {
		It("Returns a ClientFunc", func() {
			client := &http.Client{}
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache)

			result := cacheableMiddleware(client)

			Expect(reflect.TypeOf(result).Name()).To(Equal("ClientFunc"))
		})
	})

	Context("Cache ClientFunc", func() {
		It("Calls Parent Method and caches result on success", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache)
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)
			resp := &http.Response{}

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(resp, nil)
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any())

			_, _ = result.Do(req)
		})

		It("Returns and error on parent error", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache)
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("error"))

			_, err := result.Do(req)

			Expect(err).To(HaveOccurred())
		})

		It("Returns and error on parent error and does not set cache", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache)
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)

			mockCache.EXPECT().Get(gomock.Any())
			mockClient.EXPECT().Do(gomock.Any()).Return(nil, errors.New("error"))
			mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

			_, err := result.Do(req)

			Expect(err).To(HaveOccurred())
		})

		It("Does not call parent method on cache hit", func() {
			cacheableMiddleware := cacheable.NewCacheableMiddleware(mockCache)
			result := cacheableMiddleware(mockClient)
			req, _ := http.NewRequest(http.MethodGet, "localhost", nil)
			resp := http.Response{}

			mockCache.EXPECT().Get(gomock.Any()).Return(resp, true)

			_, _ = result.Do(req)
		})
	})
})
