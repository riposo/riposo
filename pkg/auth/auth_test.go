package auth_test

import (
	"testing"

	. "github.com/bsm/ginkgo"
	. "github.com/bsm/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "pkg/auth")
}
