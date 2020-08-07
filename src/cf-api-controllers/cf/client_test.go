package cf

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/mocks"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sclevine/spec"
)

func TestClientUpdateBuild(t *testing.T) {
	spec.Run(t, "TestClientUpdateBuild", func(t *testing.T, when spec.G, it spec.S) {
		const (
			guid = "guid"
		)
		var (
			client *Client
			build  model.Build
		)

		it.Before(func() {
			client = new(Client)
			client.host = "http://capi.host"
			client.restClient = new(mocks.Rest)
			client.uaaClient = new(mocks.TokenFetcher)
		})

		it.After(func() {
			mock.AssertExpectationsForObjects(t, client.restClient, client.uaaClient)
		})

		when("successfully updates", func() {
			it.Before(func() {
				build = model.Build{State: "SUCCESS"}
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				client.restClient.(*mocks.Rest).
					On("Patch", mock.Anything, mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: 200}, nil)
			})

			it("fetches a token and updates CF API server", func() {
				assert.NoError(t, client.UpdateBuild(guid, build))
				client.uaaClient.(*mocks.TokenFetcher).AssertCalled(t, "Fetch")

				raw, err := json.Marshal(build)
				assert.Empty(t, err)
				client.restClient.(*mocks.Rest).AssertCalled(t, "Patch",
					"http://capi.host/v3/builds/guid",
					"valid-token",
					bytes.NewReader(raw),
				)
			})
		})

		when("uaa client fails to fetch a token", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("", errors.New("fail"))
			})

			it("errors", func() {
				assert.Error(t, client.UpdateBuild(guid, build))
			})
		})

		when("CF API server client fails to Patch", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				client.restClient.(*mocks.Rest).
					On("Patch", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("fail"))
			})

			it("errors", func() {
				assert.Error(t, client.UpdateBuild(guid, build))
			})
		})

		when("a non-400+ status code is received", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				client.restClient.(*mocks.Rest).
					On("Patch", mock.Anything, mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: 500}, nil)
			})

			it("errors", func() {
				assert.Error(t, client.UpdateBuild(guid, build))
			})
		})
	})
}

func TestClientUpdateDroplet(t *testing.T) {
	spec.Run(t, "TestClientUpdateDroplet", func(t *testing.T, when spec.G, it spec.S) {
		const (
			guid = "guid"
		)
		var (
			client *Client
			droplet  model.Droplet
		)

		it.Before(func() {
			client = new(Client)
			client.host = "http://capi.host"
			client.restClient = new(mocks.Rest)
			client.uaaClient = new(mocks.TokenFetcher)
		})

		it.After(func() {
			mock.AssertExpectationsForObjects(t, client.restClient, client.uaaClient)
		})

		when("successfully updates", func() {
			it.Before(func() {
				droplet = model.Droplet{Image: "updated-image-reference"}
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				client.restClient.(*mocks.Rest).
					On("Patch", mock.Anything, mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: 200}, nil)
			})

			it("fetches a token and updates CF API server", func() {
				assert.NoError(t, client.UpdateDroplet(guid, droplet))
				client.uaaClient.(*mocks.TokenFetcher).AssertCalled(t, "Fetch")

				raw, err := json.Marshal(droplet)
				assert.Empty(t, err)
				client.restClient.(*mocks.Rest).AssertCalled(t, "Patch",
					"http://capi.host/v3/droplets/guid",
					"valid-token",
					bytes.NewReader(raw),
				)
			})
		})

		when("uaa client fails to fetch a token", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("", errors.New("fail"))
			})

			it("errors", func() {
				assert.Error(t, client.UpdateDroplet(guid, droplet))
			})
		})

		when("CF API server client fails to Patch", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				client.restClient.(*mocks.Rest).
					On("Patch", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("fail"))
			})

			it("errors", func() {
				assert.Error(t, client.UpdateDroplet(guid, droplet))
			})
		})

		when("a non-400+ status code is received", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				client.restClient.(*mocks.Rest).
					On("Patch", mock.Anything, mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: 500}, nil)
			})

			it("errors", func() {
				assert.Error(t, client.UpdateDroplet(guid, droplet))
			})
		})
	})
}
