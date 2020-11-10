package binary_test

import (
	"net/http/httptest"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Compare", func() {
	var testServer *httptest.Server
	BeforeEach(func() {
		testServer = newCfTestAPIServer()
	})

	AfterEach(func() {
		testServer.Close()
	})

	When("There is no difference in CF State", func() {
		It("should report no changes", func() {
			command := exec.Command(pathToBinary, "compare") // #nosec
			command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}
			writeToPipe(command, getFileContents("fixtures/cf_state/metadata.json"))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			Eventually(session).Should(gbytes.Say("No differences found between input and current state\n"))
		})
	})

	When("There is a difference between CF State", func() {
		It("should report newly added items", func() {
			command := exec.Command(pathToBinary, "compare") // #nosec
			command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}
			writeToPipe(command, getFileContents("fixtures/comparer/compare_json/current-minus-app.json"))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			sessionOutput := sanitize(string(session.Wait().Out.Contents()))
			Expect(sessionOutput).Should(Equal(sanitize(getFileContents("fixtures/comparer/compare_json/expectedOutputAdded"))))
		})

		It("should report deleted items", func() {
			command := exec.Command(pathToBinary, "compare") // #nosec
			command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}
			writeToPipe(command, getFileContents("fixtures/comparer/compare_json/current-plus-app.json"))

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
			sessionOutput := sanitize(string(session.Wait().Out.Contents()))
			expectedOutput := sanitize(getFileContents("fixtures/comparer/compare_json/expectedOutputSubtracted"))
			Expect(sessionOutput).Should(Equal(expectedOutput))
		})
	})

	When("the input JSON is invalid", func() {
		It("returns a helpful error", func() {
			command := exec.Command(pathToBinary, "compare") // #nosec
			command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}
			writeToPipe(command, "This is }not{ JSON.")

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("could not compare current CF metadata with input JSON"))
		})
	})

	When("the input JSON is not a CF-metadata JSON", func() {
		It("should report it", func() {
			command := exec.Command(pathToBinary, "compare") // #nosec
			command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}
			writeToPipe(command, `{"foo": "bar"}`)

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("could not compare current CF metadata with input JSON"))
		})
	})

	When("an unknown option is passed", func() {
		It("should report an error", func() {
			command := exec.Command(pathToBinary, "compareee") // #nosec
			command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}

			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("unknown option: compareee, Did you mean compare ?"))
		})
	})
})
