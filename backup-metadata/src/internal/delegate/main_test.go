package delegate_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/cf-for-k8s-disaster-recovery/backup-metadata/src/internal/delegate"
	"github.com/pivotal/cf-for-k8s-disaster-recovery/backup-metadata/src/internal/delegate/delegatefakes"
)

var _ = Describe("Main", func() {
	When("Reading from Stdin fails", func() {
		It("Should report an error", func() {
			const ErrorMessage = "Error Whilst reading file"
			stdIn := new(delegatefakes.FakeReader)
			stdIn.ReadReturns(0, fmt.Errorf(ErrorMessage))

			env := map[string]string{
				"CF_API":      "DUMMY_API",
				"CF_USER":     "DUMMY_USER",
				"CF_PASSWORD": "DUMMY_PASSWORD",
			}

			err := delegate.Main([]string{"", "compare"}, stdIn, env)

			Expect(err).To(MatchError(ContainSubstring(ErrorMessage)))
		})
	})
})
