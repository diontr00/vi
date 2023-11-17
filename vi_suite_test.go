package vi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vi Suite")
}
