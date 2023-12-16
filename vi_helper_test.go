package vi

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// return router helper method
func getRoute(router *vi, method string) func(path string, handler http.HandlerFunc) {
	var route func(path string, handler http.HandlerFunc)
	switch method {
	case "GET":
		route = router.GET
	case "POST":
		route = router.POST
	case "PUT":
		route = router.PUT
	case "DELETE":
		route = router.DELETE
	case "PATCH":
		route = router.PATCH
	default:
		Fail(fmt.Sprintf("Unknown method : %s", method))
	}

	return route
}

// build check ,  register handler for each path with  given method and assert the handler has been call and return correct payload
func checkSimpleResponse(url, path string, expectFail bool) {
	router := New(&Config{Banner: false})
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
		req := httptest.NewRequest(method, url, http.NoBody)
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
	router := New(&Config{Banner: false})

	h := func(w http.ResponseWriter, r *http.Request) {
		for k, v := range expectParam {
			Expect(GetParam(r, string(k))).To(Equal(v), fmt.Sprintf("param with key: %s and value: %s should be include in the context of request with path : %s", k, v, url))
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
