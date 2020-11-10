package binary_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("CF Metadata Retrieval", func() {
	It("prints an error message when client fails to initialize", func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/v2/info", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		})

		testServer := httptest.NewServer(mux)
		defer testServer.Close()

		command := interactiveShell(pathToBinary) // #nosec
		command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(1))
		Eventually(session.Err).Should(gbytes.Say(`Could not get api /v2/info: EOF`))
	})

	It("reports cloud foundry metadata", func() {
		testServer := newCfTestAPIServer()
		defer testServer.Close()

		command := interactiveShell(pathToBinary) // #nosec
		command.Env = []string{"CF_API_HOST=" + testServer.URL, "CF_CLIENT=" + cfClient, "CF_CLIENT_SECRET=" + cfClientSecret}
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)

		Expect(err).NotTo(HaveOccurred())

		cfMetadata := strings.Replace(string(session.Wait().Out.Contents()), "CF Metadata: ", "", 1)

		Eventually(session).Should(gexec.Exit(0))
		Eventually(cfMetadata).Should(MatchJSON(getFileContents("fixtures/cf_state/metadata.json")))
	})
})
