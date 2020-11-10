package binary_test

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var (
	pathToBinary   string
	cfClient       = "test-cf-user"
	cfClientSecret = "test-cf-password"
)

var _ = BeforeSuite(func() {
	var err error
	pathToBinary, err = gexec.Build("../../main.go")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func TestBinary(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Binary Suite")
}

func newCfTestAPIServer() *httptest.Server {
	var cfAPI string

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"token_endpoint": "%s"} `, cfAPI)
	})

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		bodyStr := string(body)
		authHeader := r.Header.Get("Authorization")
		expectedAuth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", cfClient, cfClientSecret)))

		if strings.Contains(authHeader, expectedAuth) &&
			strings.Contains(bodyStr, "grant_type=client_credentials") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token" : "token"}`)

			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("401 - Wrong credentials"))
	})

	mux.HandleFunc("/v2/organizations", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, getFileContents("fixtures/cf_state/organizations.json"))
	})

	mux.HandleFunc("/v2/spaces", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, getFileContents("fixtures/cf_state/spaces.json"))
	})

	mux.HandleFunc("/v2/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, getFileContents("fixtures/cf_state/users.json"))
	})

	mux.HandleFunc("/v2/apps", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, getFileContents("fixtures/cf_state/apps.json"))
	})

	testServer := httptest.NewServer(mux)
	cfAPI = testServer.URL

	return testServer
}

func getFileContents(filePath string) string {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}

	return string(fileBytes)
}

func sanitize(str string) string {
	return removeSpaces(stripAnsiColors(str))
}

func removeSpaces(str string) string {
	return strings.Replace(str, " ", "", -1)
}

const ansi = "[\u001B\u009B][[\\]()#;?]*" +
	"(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|" +
	"(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

func stripAnsiColors(str string) string {
	return re.ReplaceAllString(str, "")
}

func writeToPipe(command *exec.Cmd, contents string) io.WriteCloser {
	stdin, err := command.StdinPipe()
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	_, err = io.WriteString(stdin, contents)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	stdin.Close()

	return stdin
}

func interactiveShell(path string, arg ...string) *exec.Cmd {
	command := exec.Command(path, arg...) // #nosec
	command.Stdin, _ = os.Open("/dev/zero")

	return command
}
