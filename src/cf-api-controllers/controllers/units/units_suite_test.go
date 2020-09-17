package units_test

import (
	"testing"

	"github.com/matt-royal/biloba"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnits(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Units Suite", biloba.GoLandReporter())
}
