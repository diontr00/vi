package vi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("Test with no params", func(url, path string, expectFail bool) {
	checkSimpleResponse(url, path, expectFail)
},
	Entry("url match path", "/hello", "/hello", false),
	Entry("url not match path", "/world", "/hello", true),
)

var _ = DescribeTable("Test with params", func(url string, path string, expectParams map[matchKey]string, expectMatch bool) {
	checkResponseWithParam(url, path, expectParams, expectMatch)

},
	// Test simple param

	Entry("", "/anh", "/:name", map[matchKey]string{"name": "anh"}, true),
	Entry("", "/user/anh", "/user/:name", map[matchKey]string{"name": "anh"}, true),
	Entry("", "/user/anh/101", "/user/:name/:id", map[matchKey]string{"name": "anh", "id": "101"}, true),
	Entry("", "/video/anh/fitness", "/video/:name/:category", map[matchKey]string{"name": "anh", "category": "fitness"}, true),
	Entry("", "/video/anh/fitness", "/video/:name/:category", map[matchKey]string{"name": "anh", "category": "fitness"}, true),

	// Test param with modifier
	Entry("", "/anh", "/:name?", map[matchKey]string{"name": "anh"}, true),
	Entry("", "/user/anh", "/user/:name*", map[matchKey]string{"name": "anh"}, true),
	Entry("", "/user/anh/101", "/user/:name?/:id", map[matchKey]string{"name": "anh", "id": "101"}, true),
	Entry("", "/video/anh", "/*", nil, true),

	// Test param with regex
	Entry("", "/anh", `/{name:[\w]+}`, map[matchKey]string{"name": "anh"}, true),
	Entry("", "/user/anh", "/user/{name:[a-zA-Z]+}", map[matchKey]string{"name": "anh"}, true),
	Entry("", "/user/anh/101", "/user/:name/{id:[0-9]+}", map[matchKey]string{"name": "anh", "id": "101"}, true),

	// Other cases

	Entry("", "/anh", `/?`, nil, false),
	Entry("", "/user/anh", `/us/{name:\w+}`, nil, false),
	Entry("", "/user", "user", nil, true),
)

var _ = Describe("Test with middlewares", func() {
	var mock *handlerMock

	BeforeEach(func() {
		mock = newMock()
		mock.WrapMW("/", func(next http.HandlerFunc) http.HandlerFunc {
			return next
		})
	})

	It("Should call the register middleware", func() {

		v := New(&Config{Banner: false})
		v.Use(mock.GetMw("/"))
		mock.On("CallMock", "/").Return()
		v.GET("/", func(w http.ResponseWriter, r *http.Request) {
			return
		})

		req := httptest.NewRequest("GET", "/", http.NoBody)
		rec := httptest.NewRecorder()
		v.ServeHTTP(rec, req)
		mock.AssertCalled(GinkgoT(), "CallMock", "/")
	})

	It("Should call nested register middleware", func() {

		mock.WrapMW("/hello", func(next http.HandlerFunc) http.HandlerFunc {
			return next
		})

		mock.WrapMW("/hello/world", func(next http.HandlerFunc) http.HandlerFunc {
			return next
		})

		customNotFoundMsg := "not found handler called"
		v := New(&Config{Banner: false, NotFoundHandler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(customNotFoundMsg))
		}})

		v.Use(mock.GetMw("/"))

		sv := v.Group("/hello")
		ssv := sv.Group("world")
		sv.Use(mock.GetMw("/hello"))
		ssv.Use(mock.GetMw("/hello/world"))

		mock.On("CallMock", "/").Return()
		mock.On("CallMock", "/hello").Return()
		mock.On("CallMock", "/hello/world").Return()

		sv.GET("/hello/world", func(w http.ResponseWriter, r *http.Request) {
			return
		})

		ssv.GET("/hello/world/2", func(w http.ResponseWriter, r *http.Request) {
			return
		})

		req := httptest.NewRequest("GET", "/hello/world/2", http.NoBody)
		rec := httptest.NewRecorder()
		v.ServeHTTP(rec, req)
		mock.AssertCalled(GinkgoT(), "CallMock", "/")
		mock.AssertCalled(GinkgoT(), "CallMock", "/hello")
		mock.AssertCalled(GinkgoT(), "CallMock", "/hello/world")

		req = httptest.NewRequest("GET", "/notfound", http.NoBody)
		rec = httptest.NewRecorder()

		sv.ServeHTTP(rec, req)

		Expect(rec.Result().StatusCode).To(Equal(404))
		Expect(rec.Body.String()).To(Equal(customNotFoundMsg))

	})

	It("Test Coverage", func() {
		_ = New(&Config{Banner: true})
		r := httptest.NewRequest("GET", "/", http.NoBody)
		s := GetParam(r, "not-exit")
		Expect(s).To(BeZero())
	})
	It("Panic Case", func() {
		r := New(&Config{Banner: false})
		Ω(func() { r.Add("", "/", func(w http.ResponseWriter, r *http.Request) {}) }).Should(Panic())
		Ω(func() { r.Add("GET", "", func(w http.ResponseWriter, r *http.Request) {}) }).Should(Panic())

		Ω(func() { r.Add("GET", "/", nil) }).Should(Panic())
	})

})

var _ = Describe("Serving Static with simple file content and header", func() {
	maxAge := 2
	v := New(&Config{Banner: false})

	v.Static("/", &StaticConfig{Root: http.Dir("./.github/testdata/fs"), MaxAge: maxAge, NotFoundFile: "error/error.html"})

	cacheControl := "public, max-age=" + strconv.Itoa(maxAge)

	DescribeTable("", func(url string, expectStatusCode int, contentType string) {

		req := httptest.NewRequest("GET", url, http.NoBody)
		rec := httptest.NewRecorder()
		v.ServeHTTP(rec, req)

		Expect(rec.Result().StatusCode).To(Equal(expectStatusCode))
		Expect(rec.Header().Get("Content-Type")).To(Equal(contentType), "expect return correct Content-Type header")
		Expect(rec.Header().Get("Cache-Control")).To(Equal(cacheControl), "expect return correct Cache-Control header")

		var testFilePath string
		if rec.Result().StatusCode == 404 {
			testFilePath = "./.github/testdata/fs/error/error.html"
		} else {
			testFilePath = "./.github/testdata/fs" + url
		}

		file, err := os.Open(testFilePath)
		Expect(err).ToNot(HaveOccurred(), "[Setup] test file content %s should be readable : %v", testFilePath, err)

		stat, _ := file.Stat()
		var b = make([]byte, stat.Size())
		file.Read(b)
		Expect(string(b)).To(Equal(rec.Body.String()), "expect return correct file content")

	},
		Entry("When require index.html file", "/index.html", 200, "text/html"),
		Entry("When require style.css file", "/css/style.css", 200, "text/css"),
		Entry("When require index.js file", "/src/index.js", 200, "text/javascript"),
		Entry("When require notfound.js file", "/src/notfound.js", 404, "text/html"),
	)

})

var _ = Describe("Serving Static with prefix and next function", func() {
	maxAge := 2

	cacheControl := "public, max-age=" + strconv.Itoa(maxAge)
	v := New(&Config{Banner: false})

	nextFunc := func(w http.ResponseWriter, r *http.Request) bool {
		return r.URL.Query().Get("ignore") == "true"
	}

	v.Static("/", &StaticConfig{Root: http.Dir("./.github/testdata/fs"), MaxAge: maxAge, Next: nextFunc, Index: "index.html", Prefix: "sub"})

	DescribeTable("", func(url string, expectStatusCode int, contentType string, expectSkip bool) {

		prefix := "/sub"
		req := httptest.NewRequest("GET", url, http.NoBody)
		rec := httptest.NewRecorder()
		v.ServeHTTP(rec, req)

		if expectSkip {
			Expect(rec.Result().StatusCode).To(Equal(http.StatusNoContent))
			Expect(rec.Body.Available()).To(Equal(0))
			return
		}

		if rec.Result().StatusCode == 404 {
			Expect(rec.Body.String()).To(Equal("404 Not Found"), "expect default error message")
			return
		}

		Expect(rec.Result().StatusCode).To(Equal(expectStatusCode))
		Expect(rec.Header().Get("Content-Type")).To(Equal(contentType), "expect return correct Content-Type header")
		Expect(rec.Header().Get("Cache-Control")).To(Equal(cacheControl), "expect return correct Cache-Control header")

		testFilePath := "./.github/testdata/fs" + prefix + url

		file, err := os.Open(testFilePath)
		Expect(err).ToNot(HaveOccurred(), "[Setup] test file content %s should be readable : %v", testFilePath, err)

		stat, _ := file.Stat()
		var b = make([]byte, stat.Size())
		file.Read(b)
		Expect(string(b)).To(Equal(rec.Body.String()), "expect return correct file content")

	},
		Entry("When require index.html file", "/index.html", 200, "text/html", false),
		Entry("Expect skip", "/index.html?ignore=true", http.StatusNoContent, "", true),
		Entry("not found", "/notfound.js", 404, "text/plain", false),
	)

})

// EDGE CASES
type errorFS struct{}

func (e errorFS) Open(name string) (http.File, error) {
	return nil, errors.New("custom error for testing")
}

var _ = Describe("Edge Cases", func() {
	var v *vi
	var rec *httptest.ResponseRecorder
	BeforeEach(func() {
		v = New(&Config{Banner: false})
		rec = httptest.NewRecorder()

	})

	It("Should be panic if root is nill", func() {
		Ω(func() {
			v.Static("/", &StaticConfig{Root: nil})
		}).Should(Panic())
	})

	It("Should return internal error when request file cannot be open", func() {

		v.Static("/", &StaticConfig{Root: errorFS{}})
		req := httptest.NewRequest("GET", "/", http.NoBody)
		v.ServeHTTP(rec, req)
		Expect(rec.Result().StatusCode).To(Equal(http.StatusInternalServerError))
		Expect(rec.Body.String()).To(Equal("500 server internal error"))
	})
	It("Should return not found when receive method other then GET", func() {
		v.Static("/", &StaticConfig{Root: http.Dir("./.github/testdata/fs")})
		req := httptest.NewRequest("DELETE", "/index.html", http.NoBody)

		v.ServeHTTP(rec, req)

		Expect(rec.Result().StatusCode).To(Equal(http.StatusNotFound))
	})

	It("if file is dir , return index", func() {
		v.Static("/", &StaticConfig{Root: http.Dir("./.github/testdata/fs/")})
		req := httptest.NewRequest("GET", "/css", http.NoBody)

		v.ServeHTTP(rec, req)

		Expect(rec.Result().StatusCode).To(Equal(http.StatusOK))
		Expect(rec.Result().Header.Get("Content-Type")).To(Equal("text/html"))
	})

	It("if index file couldn't be open", func() {
		v.Static("/", &StaticConfig{Root: http.Dir("./.github/testdata/fs/"), Index: "notfoundindex.html"})
		req := httptest.NewRequest("GET", "/css", http.NoBody)

		Ω(func() {
			v.ServeHTTP(rec, req)
		}).Should(Panic())

	})

})
