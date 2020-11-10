package cfmetadata

import (
	"bytes"
	"encoding/json"

	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

type Metadata struct {
	Totals Totals       `json:"totals"`
	Orgs   []OrgSummary `json:"orgs"`
}

type Totals struct {
	Orgs   int `json:"orgs"`
	Spaces int `json:"spaces"`
	Users  int `json:"users"`
	Apps   int `json:"apps"`
}

type OrgSummary struct {
	Name   string         `json:"name"`
	Spaces []SpaceSummary `json:"spaces"`
}

type SpaceSummary struct {
	Name string       `json:"name"`
	Apps []AppSummary `json:"apps"`
}

type AppSummary struct {
	Name string `json:"name"`
}

type MetadataGetter struct {
	CfClient CfClient
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CfClient
type CfClient interface {
	Orgs() ([]Org, error)
	Spaces() ([]Space, error)
	Apps() ([]App, error)
	Users() ([]User, error)
}

func NewMetadataGetter(cfClient CfClient) (*MetadataGetter, error) {
	return &MetadataGetter{
		CfClient: cfClient,
	}, nil
}

func (mg *MetadataGetter) Execute() (*Metadata, error) {
	orgs, _ := mg.CfClient.Orgs()
	spaces, _ := mg.CfClient.Spaces()
	apps, _ := mg.CfClient.Apps()
	users, _ := mg.CfClient.Users()
	orgsSummary := mg.getAllOrgsSummary(orgs, spaces, apps)

	return &Metadata{
		Totals: Totals{
			Orgs:   len(orgs),
			Spaces: len(spaces),
			Apps:   len(apps),
			Users:  len(users),
		},
		Orgs: orgsSummary,
	}, nil
}

func (mg *MetadataGetter) getAllOrgsSummary(cfOrgs []Org, cfSpaces []Space, cfApps []App) []OrgSummary {
	appMap := mg.buildAppMap(cfApps)
	spacesMap := mg.buildSpacesMap(appMap, cfSpaces)

	orgs := []OrgSummary{}

	for _, cfOrg := range cfOrgs {
		org := OrgSummary{
			Name:   cfOrg.Name,
			Spaces: spacesMap[cfOrg.GUID],
		}

		orgs = append(orgs, org)
	}

	return orgs
}

func (mg *MetadataGetter) buildAppMap(cfApps []App) map[string][]AppSummary {
	appMap := make(map[string][]AppSummary)

	for _, cfApp := range cfApps {
		spaceGUID := cfApp.SpaceGUID
		appMap[spaceGUID] = append(appMap[spaceGUID], AppSummary{Name: cfApp.Name})
	}

	return appMap
}

func (mg *MetadataGetter) buildSpacesMap(appMap map[string][]AppSummary, cfSpaces []Space) map[string][]SpaceSummary {
	spaceMap := make(map[string][]SpaceSummary)

	for _, cfSpace := range cfSpaces {
		orgGUID := cfSpace.OrgGUID
		spaceMap[orgGUID] = append(spaceMap[orgGUID], SpaceSummary{
			Name: cfSpace.Name, Apps: appMap[cfSpace.GUID],
		})
	}

	return spaceMap
}

func Compare(leftMetadataStr []byte, rightMetadata Metadata) (string, error) {
	var leftMetadata Metadata

	dec := json.NewDecoder(bytes.NewReader(leftMetadataStr))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&leftMetadata); err != nil {
		return "", err
	}

	return compare(toStr(leftMetadata), toStr(rightMetadata)), nil
}

func toStr(metadata Metadata) []byte {
	ret, _ := json.Marshal(metadata)

	return ret
}

func compare(leftJSON []byte, rightJSON []byte) string {
	differ := diff.New()
	d, _ := differ.Compare(leftJSON, rightJSON)

	if !d.Modified() {
		return ""
	}

	var aJSON map[string]interface{}
	_ = json.Unmarshal(leftJSON, &aJSON)

	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
		Coloring:       true,
	}

	asciiFormatter := formatter.NewAsciiFormatter(aJSON, config)
	diffString, _ := asciiFormatter.Format(d)

	return diffString
}
