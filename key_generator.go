package cacheable

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

// GenerateKeyHash A generic method for creating a string cache key from an http.Request object
func GenerateKeyHash(r *http.Request) string {
	s := sha1.New()
	s.Write([]byte(r.Method))
	s.Write([]byte(r.URL.Host))
	s.Write([]byte(r.URL.Path))

	paramList := getSortedParams(r)
	for _, value := range paramList {
		s.Write([]byte(value))
	}

	headersList := getSortedHeaders(r)
	for _, value := range headersList {
		s.Write([]byte(value))
	}

	if r.Body != nil {
		defer r.Body.Close()
		body, _ := ioutil.ReadAll(r.Body)
		if len(body) > 0 {
			s.Write(body)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	return base64.StdEncoding.EncodeToString(s.Sum(nil))
}

func getSortedHeaders(r *http.Request) []string {
	headersList := make([]string, 0, len(r.Header))
	for key, value := range r.Header {
		headersList = append(headersList, key+"="+strings.Join(value, ","))
	}
	if !sort.StringsAreSorted(headersList) {
		sort.Strings(headersList)
	}
	return headersList
}

func getSortedParams(r *http.Request) []string {
	paramList := make([]string, 0, len(r.URL.Query()))
	for key, value := range r.URL.Query() {
		paramList = append(paramList, key+"="+strings.Join(value, ","))
	}
	if !sort.StringsAreSorted(paramList) {
		sort.Strings(paramList)
	}
	return paramList
}
