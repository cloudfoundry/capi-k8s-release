package helpers_test

import (
	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata-generator/internal/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Env Helper", func() {
	It("maps os.Environ() to a dictionary", func() {
		environ := []string{
			`FIRST_KEY=foo`,
			"SECOND_KEY=bar",
			"VALUE_WITH_EQUAKS=this=this",
		}

		env := helpers.EnvironToMap(environ)

		Expect(env["FIRST_KEY"]).To(Equal("foo"))
		Expect(env["SECOND_KEY"]).To(Equal("bar"))
		Expect(env["VALUE_WITH_EQUAKS"]).To(Equal("this=this"))
		Expect(len(env)).To(Equal(3))
	})

	It("maps empty os.Environ() to an empty dictionary ", func() {
		environ := []string{}
		env := helpers.EnvironToMap(environ)

		Expect(env).NotTo(BeNil())
		Expect(len(env)).To(Equal(0))
	})
})
