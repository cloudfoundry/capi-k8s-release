package e2e_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata-generator/internal/cfmetadata"
)

var _ = Describe("e2e", func() {
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
	Expect(err).Should(BeNil())
}

func getCFMetadataFromBackupLogs(backupName string) string {
	command := "velero backup logs " + backupName + " | grep -oP '(?<=CF Metadata: ){.*}(?=\\\\n)' | xargs echo"

	backupLogs, err := exec.Command("bash", "-c", command).Output()
	Expect(err).Should(BeNil())

	return string(backupLogs)
}
