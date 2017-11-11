package tlb_test

import (
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTLB(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TLB Suite")
}
