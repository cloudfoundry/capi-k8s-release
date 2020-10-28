package delegate_test

import (
	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata/internal/delegate"
	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata/internal/delegate/delegatefakes"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	When("Reading from Stdin fails", func() {
		It("Should report an error", func() {
			const ErrorMessage = "Error Whilst reading file"
			stdIn := new(delegatefakes.FakeReader)
			stdIn.ReadReturns(0, fmt.Errorf(ErrorMessage))

			env := map[string]string{
				"CF_API_HOST":      "api.cf.example.com",
				"CF_CLIENT":        "fake-uaa-client",
				"CF_CLIENT_SECRET": "fake-uaa-client-secret",
			}

			err := delegate.Main([]string{"", "compare"}, stdIn, env)

			Expect(err).To(MatchError(ContainSubstring(ErrorMessage)))
		})
	})

	DescribeTable("when environment variables are missing",
		func(env map[string]string) {
			const errorMessage = "Environment Variable is not set"
			stdIn := new(delegatefakes.FakeReader)
			stdIn.ReadReturns(0, nil)

			err := delegate.Main([]string{"", "compare"}, stdIn, env)
			Expect(err).To(MatchError(ContainSubstring(errorMessage)))
		},
		Entry("CF_API_HOST is missing", map[string]string{"CF_CLIENT": "fake-uaa-client", "CF_CLIENT_SECRET": "fake-uaa-client-secret"}),
		Entry("CF_CLIENT is missing", map[string]string{"CF_API_HOST": "api.cf.example.com", "CF_CLIENT_SECRET": "fake-uaa-client-secret"}),
		Entry("CF_CLIENT_SECRET is missing", map[string]string{"CF_API_HOST": "api.cf.example.com", "CF_CLIENT": "fake-uaa-client"}),
	)
})
