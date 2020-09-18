package cfg_test

import (
	"github.com/matt-royal/biloba"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCfg(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t, "Cfg Suite", biloba.GoLandReporter())
}
