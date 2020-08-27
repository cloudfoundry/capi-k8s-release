package helpers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal/cf-for-k8s-disaster-recovery/backup-metadata/src/internal/helpers"
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
