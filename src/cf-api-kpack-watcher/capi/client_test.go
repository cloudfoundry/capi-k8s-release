package capi

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi/mocks"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher/capi_model"

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
			status capi_model.Build
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
				status = capi_model.Build{State: "SUCCESS"}
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				client.restClient.(*mocks.Rest).
					On("Patch", mock.Anything, mock.Anything, mock.Anything).
					Return(&http.Response{StatusCode: 200}, nil)
			})

			it("fetches a token and updates CF API server", func() {
				assert.NoError(t, client.UpdateBuild(guid, status))
				client.uaaClient.(*mocks.TokenFetcher).AssertCalled(t, "Fetch")
				client.restClient.(*mocks.Rest).AssertCalled(t, "Patch",
					"http://capi.host/v3/builds/guid",
					"valid-token",
					bytes.NewReader(status.ToJSON()),
				)
			})
		})

		when("uaa client fails to fetch a token", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("", errors.New("fail"))
			})

			it("errors", func() {
				assert.Error(t, client.UpdateBuild(guid, status))
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
				assert.Error(t, client.UpdateBuild(guid, status))
			})
		})
	})
}
