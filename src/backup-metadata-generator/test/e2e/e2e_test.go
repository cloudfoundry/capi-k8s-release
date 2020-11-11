package e2e_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata-generator/internal/cfmetadata"
)

var _ = Describe("e2e", func() {
	var cfMetadataFile *os.File

	AfterEach(func() {
		if cfMetadataFile != nil {
			os.Remove(cfMetadataFile.Name())
		}
	})

	It("Assert CF metadata structure", func() {
		backupName := generateRandomName()
		createBackup(backupName)
		cfMetadataJSON := getCFMetadataFromBackupLogs(backupName)

		var metadata cfmetadata.Metadata
		err := json.Unmarshal([]byte(cfMetadataJSON), &metadata)
		Expect(err).NotTo(HaveOccurred())

		Expect(metadata.Totals.Orgs).Should(BeNumerically(">=", 1))
		Expect(metadata.Totals.Spaces).Should(BeNumerically(">=", 0))
		Expect(metadata.Totals.Users).Should(BeNumerically(">=", 0))
		Expect(metadata.Totals.Apps).Should(BeNumerically(">=", 0))
		Expect(len(metadata.Orgs)).Should(BeNumerically(">=", 1))

		cfMetadataFile = tmpOutputFile(cfMetadataJSON)
		comparison := compareBackup(cfMetadataFile.Name())
		Expect(comparison).To(MatchRegexp("No differences found|totals"))
	})
})

func generateRandomName() string {
	return "backup-" + random()
}

func random() string {
	rand.Seed(time.Now().UnixNano())

	return strconv.Itoa(rand.Intn(1000))
}

func createBackup(backupName string) {
	fmt.Printf("Creating velero backup: %s \n", backupName)

	_, err := exec.Command("velero", "backup", "create", "--wait", backupName).Output()
	Expect(err).NotTo(HaveOccurred())
}

func getCFMetadataFromBackupLogs(backupName string) string {
	command := "../../bin/get-backup-metadata.sh " + backupName + " /dev/stdout"

	backupLogs, err := exec.Command("bash", "-c", command).Output()
	Expect(err).NotTo(HaveOccurred())

	return string(backupLogs)
}

func compareBackup(filePath string) string {
	command := fmt.Sprintf("../../bin/compare-backup-metadata.sh %s cf-system", filePath)

	comparison, err := exec.Command("bash", "-c", command).CombinedOutput()
	Expect(err).NotTo(HaveOccurred())

	return string(comparison)
}

func tmpOutputFile(content string) *os.File {
	file, err := ioutil.TempFile(os.TempDir(), "backup-e2e")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	_, err = file.WriteString(content)
	Expect(err).NotTo(HaveOccurred())

	return file
}
