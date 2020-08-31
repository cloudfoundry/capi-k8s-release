package cfmetadata

const UsersResponse = `{
	"resources": [{
			"entity": {
				"username": "Elon Musk"
			}
		},
		{
			"entity": {
				"username": "Elon Musk II"
			}
		}
	]}
`

const SpacesResponse = `{
	"resources": [{
			"metadata": {
				"guid": "space-x-id"
			},
			"entity": {
				"name": "space-x",
				"organization_guid": "org-1-id"
			}
		},
		{
			"metadata": {
				"guid": "space-y-id"
			},
			"entity": {
				"name": "space-y",
				"organization_guid": "org-1-id"
			}
		}
	]}
`

const AppsResponse = `{
	"resources": [{
			"entity": {
				"name": "test-app-1",
				"space_guid": "space-x-id"
			}
		}
	]}
`

const OrgsResponse = `{
	"resources": [{
			"metadata": {
				"guid": "org-1-id"
			},
			"entity": {
				"name": "org-1"
			}
		},
		{
			"metadata": {
				"guid": "org-2-id"
			},
			"entity": {
				"name": "org-2"
			}
		},
		{
			"metadata": {
				"guid": "org-3-id"
			},
			"entity": {
				"name": "org-3"
			}
		}
	]}
`
