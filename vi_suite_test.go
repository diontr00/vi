package vi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/stretchr/testify/mock"
)

var avaibleRoute = [...]method{MethodGet, MethodPost, MethodPut, MethodDelete, MethodPatch}

func TestVi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vi Suite")
}

type handlerMock struct {
	mock.Mock
	handler    http.HandlerFunc
	middleware map[string]middleware
}

func (m *handlerMock) WrapMW(prefix string, mw middleware) {
	m.middleware[prefix] = func(next http.HandlerFunc) http.HandlerFunc {
		m.CallMock(prefix)
		return next
	}
}

func (m *handlerMock) GetMw(prefix string) middleware {
	return m.middleware[prefix]
}

// Return wrapped handlerFunc that call mock and handler function
func (m *handlerMock) Wrap(path string, handler http.HandlerFunc) {
	m.handler = func(w http.ResponseWriter, r *http.Request) {
		m.CallMock(path)

		handler(w, r)
	}
}

func (m *handlerMock) GetHandler() http.HandlerFunc {
	return m.handler
}

func (m *handlerMock) CallMock(path string) {
	m.Called(path)
}

func (m *handlerMock) Reset() {
	m.Mock = mock.Mock{}
	m.handler = nil
}

func setupMock(h http.HandlerFunc, path string) *handlerMock {
	m := new(handlerMock)
	m.On("CallMock", path).Return()

	m.Wrap(path, h)
	return m
}

func newMock() *handlerMock {
	return &handlerMock{
		middleware: make(map[string]middleware),
	}
}
