package cfmetadata_test

import (
	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata-generator/internal/cfmetadata"
	"code.cloudfoundry.org/capi-k8s-release/src/backup-metadata-generator/internal/cfmetadata/cfmetadatafakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Metadata", func() {
	var (
		mg           *cfmetadata.MetadataGetter
		fakeCfClient *cfmetadatafakes.FakeCfClient
	)

	BeforeEach(func() {
		fakeCfClient = new(cfmetadatafakes.FakeCfClient)
		mg, _ = cfmetadata.NewMetadataGetter(fakeCfClient)
	})

	Describe("#Execute", func() {
		Context("given orgs exist", func() {
			It("returns expected structure", func() {
				fakeCfClient.OrgsReturns([]cfmetadata.Org{{Name: "test"}}, nil)

				m, err := mg.Execute()
				Expect(m.Orgs).Should(HaveLen(1))
				Expect(err).ShouldNot(HaveOccurred())
			})

			Context("given spaces exist", func() {
				It("returns expected structure", func() {
					fakeCfClient.SpacesReturns([]cfmetadata.Space{{Name: "test-space", OrgGUID: "org-1-id"}}, nil)
					fakeCfClient.OrgsReturns([]cfmetadata.Org{{Name: "test", GUID: "org-1-id"}}, nil)

					m, err := mg.Execute()
					Expect(m.Orgs[0].Spaces).Should(HaveLen(1))
					Expect(err).ShouldNot(HaveOccurred())
				})
			})
		})
	})

	Describe("#Compare", func() {
		Context("given a empty Json", func() {
			It("should return empty string when there is no difference", func() {
				diff, err := cfmetadata.Compare([]byte("{}"), cfmetadata.Metadata{})
				Expect(err).NotTo(HaveOccurred())

				const noDiff = ""
				Expect(diff).To(Equal(noDiff))
			})
		})

		Context("given a minimal valid Json", func() {
			It("should return a diff and no error", func() {
				actualDiff, err := cfmetadata.Compare([]byte(`{"totals": {"orgs": 1}}`), cfmetadata.Metadata{})
				Expect(err).NotTo(HaveOccurred())

				Expect(actualDiff).To(MatchRegexp(`- *"orgs":[ ]*1`))
				Expect(actualDiff).To(MatchRegexp(`\+ *"orgs":[ ]*0`))
				Expect(countChar("+", actualDiff)).To(Equal(1))
				Expect(countChar("-", actualDiff)).To(Equal(1))
			})
		})

		Context("given a minimal invalid Json", func() {
			It("should return an error", func() {
				_, err := cfmetadata.Compare([]byte(`{"foo": []}`), cfmetadata.Metadata{})
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

func countChar(ch string, str string) int {
	chBytes := []byte(ch)
	ExpectWithOffset(1, chBytes).To(HaveLen(1), "Expected single byte character, but found: "+ch)

	count := 0
	strBytes := []byte(str)

	for _, b := range strBytes {
		if b == chBytes[0] {
			count++
		}
	}

	return count
}
