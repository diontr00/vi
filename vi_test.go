package vi

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("Test with no params", func(url, path string, expectFail bool) {
	checkSimpleResponse(url, path, expectFail)
},
	Entry("url match path", "/hello", "/hello", false),
	Entry("url not match path", "/world", "/hello", true),
	Entry("path empty should be replace with /", "/", "", false),
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
	Entry("", "/", "", nil, true),
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

		v := New(&Config{banner: false})
		v.Use(mock.GetMw("/"))
		mock.On("CallMock", "/").Return()
		v.GET("/", func(w http.ResponseWriter, r *http.Request) {
			return
		})

		req := httptest.NewRequest(string(MethodGet), "/", http.NoBody)
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

		v := New(&Config{banner: false})

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

		req := httptest.NewRequest(string(MethodGet), "/hello/world/2", http.NoBody)
		rec := httptest.NewRecorder()
		v.ServeHTTP(rec, req)
		mock.AssertCalled(GinkgoT(), "CallMock", "/")
		mock.AssertCalled(GinkgoT(), "CallMock", "/hello")
		mock.AssertCalled(GinkgoT(), "CallMock", "/hello/world")

	})

	It("", func() {
		_ = New(&Config{banner: true})

	})

})

// return router helper method
func getRoute(router *vi, method method) func(path string, handler http.HandlerFunc) {
	var route func(path string, handler http.HandlerFunc)
	switch method {
	case MethodGet:
		route = router.GET
	case MethodPost:
		route = router.POST
	case MethodPut:
		route = router.PUT
	case MethodDelete:
		route = router.DELETE
	case MethodPatch:
		route = router.PATCH
	default:
		Fail(fmt.Sprintf("Unknown method : %s", method))
	}

	return route
}

// build check ,  register handler for each path with  given method and assert the handler has been call and return correct payload
func checkSimpleResponse(url, path string, expectFail bool) {
	router := New(&Config{banner: false})
	//  validate trees == nil  case
	router.trees = nil

	h := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.URL.Path)
	}

	mock := setupMock(h, path)
	handler := mock.GetHandler()

	for _, method := range avaibleRoute {
		route := getRoute(router, method)
		route(path, handler)
	}

	for _, method := range avaibleRoute {
		req := httptest.NewRequest(string(method), url, http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if expectFail {
			Expect(rec.Result().StatusCode).To(Equal(404), "should return http error not found for url :%s", path)
			return
		}

		expectBody := path
		if expectBody == "" {
			expectBody = "/"
		}

		Expect(rec.Body.String()).To(Equal(expectBody), fmt.Sprintf("Expect receive %s as the body for path %s , with method %s but got %s", expectBody, path, method, rec.Body.String()))

		mock.AssertCalled(GinkgoT(), "CallMock", path)
	}
}

// check whether handler register for path  match url and handle correctly with param
func checkResponseWithParam(url, path string, expectParam map[matchKey]string, expectMatch bool) {
	router := New(&Config{banner: false})

	h := func(w http.ResponseWriter, r *http.Request) {
		for k, v := range expectParam {
			Expect(r.Context().Value(k)).To(Equal(v), fmt.Sprintf("param with key: %s and value: %s should be include in the context of request with path : %s", k, v, url))
		}

		fmt.Fprintf(w, "%s with param: %v", r.URL.Path, expectParam)
	}

	m := setupMock(h, path)

	handler := m.GetHandler()

	for _, method := range avaibleRoute {
		route := getRoute(router, method)
		route(path, handler)
	}

	for _, method := range avaibleRoute {
		req := httptest.NewRequest(string(method), url, http.NoBody)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		if !expectMatch {
			Expect(rec.Result().StatusCode).To(Equal(404), "should return http error not found for url :%s", url)

			return
		}

		expectBody := fmt.Sprintf("%s with param: %v", url, expectParam)
		Expect(rec.Body.String()).To(Equal(expectBody), fmt.Sprintf("Expect receive %s as the body for path %s , with method %s", expectBody, path, method))

		m.AssertCalled(GinkgoT(), "CallMock", path)
	}
}
