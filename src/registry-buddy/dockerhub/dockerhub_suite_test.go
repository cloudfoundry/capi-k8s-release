package dockerhub_test

import (
	"github.com/matt-royal/biloba"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDockerhub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Dockerhub Suite", biloba.GoLandReporter())
}
