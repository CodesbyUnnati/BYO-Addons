package controller

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAddonSetController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AddonSet Controller Suite")
}
