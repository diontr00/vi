package vi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"net/http"

	"github.com/stretchr/testify/mock"
)

var avaibleRoute = [...]method{MethodGet, MethodPost, MethodPut, MethodDelete}

func TestVi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vi Suite")
}

type handlerMock struct {
	mock.Mock
	handler http.HandlerFunc
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
