package cacheable_test

import (
	"bytes"
	cacheable "github.com/TaylorOno/http-cacheable/v2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("Util", func() {
	Context("GenerateKeyHash", func() {
		It("Creates a hashkey from a http.Request", func() {
			req, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			result := cacheable.GenerateKeyHash(req)
			Expect(result).To(Equal("I/AgyjY2Ea28TA3vfYAZSp58e4Q="))
		})

		It("Creates a unique hashkey based on host", func() {
			req1, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			req2, _ := http.NewRequest(http.MethodGet, "https://localhost2", nil)
			result1 := cacheable.GenerateKeyHash(req1)
			result2 := cacheable.GenerateKeyHash(req2)
			Expect(result1).ToNot(Equal(result2))
		})

		It("Creates a unique hashkey based on method", func() {
			req1, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			req2, _ := http.NewRequest(http.MethodPost, "https://localhost", nil)
			result1 := cacheable.GenerateKeyHash(req1)
			result2 := cacheable.GenerateKeyHash(req2)
			Expect(result1).ToNot(Equal(result2))
		})

		It("Creates a unique hashkey based on path", func() {
			req1, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			req2, _ := http.NewRequest(http.MethodGet, "https://localhost/path", nil)
			result1 := cacheable.GenerateKeyHash(req1)
			result2 := cacheable.GenerateKeyHash(req2)
			Expect(result1).ToNot(Equal(result2))
		})

		It("Creates a unique hashkey based on queryParam", func() {
			req1, _ := http.NewRequest(http.MethodGet, "https://localhost?id=1", nil)
			req2, _ := http.NewRequest(http.MethodGet, "https://localhost?id=2", nil)
			result1 := cacheable.GenerateKeyHash(req1)
			result2 := cacheable.GenerateKeyHash(req2)
			Expect(result1).ToNot(Equal(result2))
		})

		It("Creates the same hashkey reguardless of queryParam order", func() {
			req1, _ := http.NewRequest(http.MethodGet, "https://localhost?id=1&limit=1", nil)
			req2, _ := http.NewRequest(http.MethodGet, "https://localhost?limit=1&id=1", nil)
			result1 := cacheable.GenerateKeyHash(req1)
			result2 := cacheable.GenerateKeyHash(req2)
			Expect(result1).To(Equal(result2))
		})

		It("Creates a uniq hashkey based on headerParam", func() {
			req1, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			req1.Header.Set("User-Agent", "FireFox")
			req2, _ := http.NewRequest(http.MethodGet, "https://localhost", nil)
			req2.Header.Set("User-Agent", "Chrome")
			result1 := cacheable.GenerateKeyHash(req1)
			result2 := cacheable.GenerateKeyHash(req2)
			Expect(result1).ToNot(Equal(result2))
		})

		It("Creates a uniq hashkey based on request body", func() {
			req1, _ := http.NewRequest(http.MethodPost, "https://localhost", bytes.NewReader([]byte(`{"limit":1}`)))
			req2, _ := http.NewRequest(http.MethodPost, "https://localhost", bytes.NewReader([]byte(`{"limit":2}`)))
			result1 := cacheable.GenerateKeyHash(req1)
			result2 := cacheable.GenerateKeyHash(req2)
			Expect(result1).ToNot(Equal(result2))
		})
	})
})
