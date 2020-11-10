package cfmetadata

import (
	"github.com/cloudfoundry-community/go-cfclient"
)

type Org struct {
	Name   string  `json:"name"`
	Spaces []Space `json:"spaces"`
	GUID   string  `json:"guid"`
}

type Space struct {
	Name    string `json:"name"`
	Apps    []App  `json:"apps"`
	GUID    string `json:"guid"`
	OrgGUID string `json:"org_guid"`
}

type App struct {
	Name      string `json:"name"`
	SpaceGUID string `json:"space_guid"`
}

type User struct {
	Name string `json:"name"`
}

type Client struct {
	CfClient cfclient.CloudFoundryClient
}

func NewClient(apiAddress string, clientID string, clientSecret string) (*Client, error) {
	conf := &cfclient.Config{
		ApiAddress:        apiAddress,
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		SkipSslValidation: true,
	}

	cfClient, err := cfclient.NewClient(conf)

	return &Client{
		CfClient: cfClient,
	}, err
}

func (c *Client) Orgs() ([]Org, error) {
	cfOrgs, err := c.CfClient.ListOrgs()
	if err != nil {
		return []Org{}, err
	}

	orgs := make([]Org, len(cfOrgs))

	for i, cfOrg := range cfOrgs {
		orgs[i] = Org{Name: cfOrg.Name, GUID: cfOrg.Guid}
	}

	return orgs, nil
}

func (c *Client) Spaces() ([]Space, error) {
	cfSpaces, err := c.CfClient.ListSpaces()
	if err != nil {
		return []Space{}, err
	}

	spaces := make([]Space, len(cfSpaces))

	for i, cfSpace := range cfSpaces {
		spaces[i] = Space{Name: cfSpace.Name, GUID: cfSpace.Guid, OrgGUID: cfSpace.OrganizationGuid}
	}

	return spaces, nil
}

func (c *Client) Apps() ([]App, error) {
	cfApps, err := c.CfClient.ListApps()
	if err != nil {
		return []App{}, err
	}

	apps := make([]App, len(cfApps))

	for i, cfApp := range cfApps {
		apps[i] = App{Name: cfApp.Name, SpaceGUID: cfApp.SpaceGuid}
	}

	return apps, nil
}

func (c *Client) Users() ([]User, error) {
	cfUsers, err := c.CfClient.ListUsers()
	if err != nil {
		return []User{}, err
	}

	users := make([]User, len(cfUsers))

	for i, cfUser := range cfUsers {
		users[i] = User{Name: cfUser.Username}
	}

	return users, nil
}
