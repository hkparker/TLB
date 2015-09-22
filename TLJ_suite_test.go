package tlj_test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//. "github.com/hkparker/TLJ"
)

func TestTLJ(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TLJ Suite")
}
