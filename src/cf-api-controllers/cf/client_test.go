package cf

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/mocks"
	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"

	"github.com/stretchr/testify/mock"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

func TestClientUpdateBuild(t *testing.T) {
	spec.Run(t, "TestClientUpdateBuild", func(t *testing.T, when spec.G, it spec.S) {
		const (
			guid = "guid"
		)
		var (
			g      = NewWithT(t)
			client *Client
			build  model.Build
		)

		it.Before(func() {
			client = NewClient("http://capi.host", new(mocks.Rest), new(mocks.TokenFetcher))
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
				g.Expect(client.UpdateBuild(guid, build)).To(Succeed())
				client.uaaClient.(*mocks.TokenFetcher).AssertCalled(t, "Fetch")

				raw, err := json.Marshal(build)
				g.Expect(err).To(BeNil())
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
				g.Expect(client.UpdateBuild(guid, build)).ToNot(Succeed())
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
				g.Expect(client.UpdateBuild(guid, build)).ToNot(Succeed())
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
				g.Expect(client.UpdateBuild(guid, build)).ToNot(Succeed())
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
			g       = NewWithT(t)
			client  *Client
			droplet model.Droplet
		)

		it.Before(func() {
			client = NewClient("http://capi.host", new(mocks.Rest), new(mocks.TokenFetcher))
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
				g.Expect(client.UpdateDroplet(guid, droplet)).To(Succeed())
				client.uaaClient.(*mocks.TokenFetcher).AssertCalled(t, "Fetch")

				raw, err := json.Marshal(droplet)
				g.Expect(err).To(BeNil())
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
				g.Expect(client.UpdateDroplet(guid, droplet)).ToNot(Succeed())
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
				g.Expect(client.UpdateDroplet(guid, droplet)).ToNot(Succeed())
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
				g.Expect(client.UpdateDroplet(guid, droplet)).ToNot(Succeed())
			})
		})
	})
}

func TestClientListRoutes(t *testing.T) {
	spec.Run(t, "TestClientListRoutes", func(t *testing.T, when spec.G, it spec.S) {
		const (
			guid = "guid"
		)
		// TODO: remove `g` and `gg` after moving to ginkgo
		var (
			g               = NewWithT(t)
			gg              = ghttp.NewGHTTPWithGomega(g)
			client          *Client
			fakeCFAPIServer *ghttp.Server
		)

		it.Before(func() {
			fakeCFAPIServer = ghttp.NewServer()

			client = NewClient(fakeCFAPIServer.URL(), new(mocks.Rest), new(mocks.TokenFetcher))
		})

		it.After(func() {
			fakeCFAPIServer.Close()
			mock.AssertExpectationsForObjects(t, client.restClient, client.uaaClient)
		})

		when("CF API is operating normally", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						gg.VerifyRequest("GET", "/v3/routes"),
						gg.RespondWith(200, `{
  "pagination": {
    "total_results": 1,
    "total_pages": 1,
    "first": {
      "href": "https://api.example.org/v3/routes?page=1&per_page=5000"
    },
    "last": {
      "href": "https://api.example.org/v3/routes?page=1&per_page=5000"
    },
    "next": null,
    "previous": null
  },
  "resources": [
    {
      "guid": "cbad697f-cac1-48f4-9017-ac08f39dfb31",
      "protocol": "http",
      "created_at": "2019-05-10T17:17:48Z",
      "updated_at": "2019-05-10T17:17:48Z",
      "host": "a-hostname",
      "path": "/some_path",
      "url": "a-hostname.a-domain.com/some_path",
      "destinations": [
        {
          "guid": "385bf117-17f5-4689-8c5c-08c6cc821fed",
          "app": {
            "guid": "0a6636b5-7fc4-44d8-8752-0db3e40b35a5",
            "process": {
              "type": "web"
            }
          },
          "weight": null,
          "port": 8080
        },
        {
          "guid": "27e96a3b-5bcf-49ed-8048-351e0be23e6f",
          "app": {
            "guid": "f61e59fa-2121-4217-8c7b-15bfd75baf25",
            "process": {
              "type": "web"
            }
          },
          "weight": null,
          "port": 8080
        }
      ],
      "metadata": {
        "labels": {},
        "annotations": {}
      },
      "relationships": {
        "space": {
          "data": {
            "guid": "885a8cb3-c07b-4856-b448-eeb10bf36236"
          }
        },
        "domain": {
          "data": {
            "guid": "0b5f3633-194c-42d2-9408-972366617e0e"
          }
        }
      },
      "links": {
        "self": {
          "href": "https://api.example.org/v3/routes/cbad697f-cac1-48f4-9017-ac08f39dfb31"
        },
        "space": {
          "href": "https://api.example.org/v3/spaces/885a8cb3-c07b-4856-b448-eeb10bf36236"
        },
        "domain": {
          "href": "https://api.example.org/v3/domains/0b5f3633-194c-42d2-9408-972366617e0e"
        },
        "destinations": {
          "href": "https://api.example.org/v3/routes/cbad697f-cac1-48f4-9017-ac08f39dfb31/destinations"
        }
      }
    }
  ]
}`),
					),
				)
			})

			it("returns a list of routes", func() {
				routes, err := client.ListRoutes()

				g.Expect(err).To(BeNil())
				g.Expect(routes).ToNot(BeEmpty())
				g.Expect(routes).To(HaveLen(1))

				g.Expect(routes[0].GUID).To(Equal("cbad697f-cac1-48f4-9017-ac08f39dfb31"))
				g.Expect(routes[0].Host).To(Equal("a-hostname"))
				g.Expect(routes[0].Path).To(Equal("/some_path"))
				g.Expect(routes[0].URL).To(Equal("a-hostname.a-domain.com/some_path"))

				g.Expect(routes[0].Destinations).To(HaveLen(2))
				g.Expect(routes[0].Destinations[0].GUID).To(Equal("385bf117-17f5-4689-8c5c-08c6cc821fed"))
				g.Expect(routes[0].Destinations[0].Port).To(Equal(8080))
				g.Expect(routes[0].Destinations[0].Weight).To(BeNil())
				g.Expect(routes[0].Destinations[0].App.GUID).To(Equal("0a6636b5-7fc4-44d8-8752-0db3e40b35a5"))
				g.Expect(routes[0].Destinations[0].App.Process.Type).To(Equal("web"))

				g.Expect(routes[0].Relationships).To(HaveKeyWithValue("space", model.Relationship{Data: model.RelationshipData{GUID: "885a8cb3-c07b-4856-b448-eeb10bf36236"}}))
				g.Expect(routes[0].Relationships).To(HaveKeyWithValue("domain", model.Relationship{Data: model.RelationshipData{GUID: "0b5f3633-194c-42d2-9408-972366617e0e"}}))
			})
		})

		when("CF API is down", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.Close()
			})

			it("returns a meaningful error", func() {
				_, err := client.ListRoutes()

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("connection refused"))
			})
		})

		when("CF API returns a non-200 status code", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						gg.VerifyRequest("GET", "/v3/routes"),
						gg.RespondWith(418, ""),
					),
				)
			})

			it("returns a meaningful error", func() {
				_, err := client.ListRoutes()

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("failed to list routes, received status: 418"))
			})
		})

		when("CF API returns an unexpected JSON response", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						gg.VerifyRequest("GET", "/v3/routes"),
						gg.RespondWith(200, `{`),
					),
				)
			})

			it("returns a meaningful error", func() {
				_, err := client.ListRoutes()

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("failed to deserialize response from CF API"))
			})
		})

		when("uaa client fails to fetch a token", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("", errors.New("uaa-fail"))
			})

			it("errors", func() {
				_, err := client.ListRoutes()

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("uaa-fail"))
			})
		})
	})
}

func TestClientGetSpace(t *testing.T) {
	spec.Run(t, "TestClientGetSpace", func(t *testing.T, when spec.G, it spec.S) {
		// TODO: remove `g` and `gg` after moving to ginkgo
		var (
			g               = NewWithT(t)
			gg              = ghttp.NewGHTTPWithGomega(g)
			client          *Client
			fakeCFAPIServer *ghttp.Server
		)

		it.Before(func() {
			fakeCFAPIServer = ghttp.NewServer()

			client = NewClient(fakeCFAPIServer.URL(), new(mocks.Rest), new(mocks.TokenFetcher))
		})

		it.After(func() {
			fakeCFAPIServer.Close()
			mock.AssertExpectationsForObjects(t, client.restClient, client.uaaClient)
		})

		when("CF API is operating normally", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						gg.VerifyRequest("GET", "/v3/spaces/885735b5-aea4-4cf5-8e44-961af0e41920"),
						gg.RespondWith(200, `{
  "guid": "885735b5-aea4-4cf5-8e44-961af0e41920",
  "created_at": "2017-02-01T01:33:58Z",
  "updated_at": "2017-02-01T01:33:58Z",
  "name": "my-space",
  "relationships": {
    "organization": {
      "data": {
        "guid": "e00705b9-7b42-4561-ae97-2520399d2133"
      }
    },
    "quota": {
      "data": null
    }
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/spaces/885735b5-aea4-4cf5-8e44-961af0e41920"
    },
    "features": {
      "href": "https://api.example.org/v3/spaces/885735b5-aea4-4cf5-8e44-961af0e41920/features"
    },
    "organization": {
      "href": "https://api.example.org/v3/organizations/e00705b9-7b42-4561-ae97-2520399d2133"
    },
    "apply_manifest": {
      "href": "https://api.example.org/v3/spaces/885735b5-aea4-4cf5-8e44-961af0e41920/actions/apply_manifest",
      "method": "POST"
    }
  },
  "metadata": {
    "labels": {},
    "annotations": {}
  }
}`),
					),
				)
			})

			it("returns an expected space object", func() {
				space, err := client.GetSpace("885735b5-aea4-4cf5-8e44-961af0e41920")

				g.Expect(err).To(BeNil())
				g.Expect(space).ToNot(BeNil())

				g.Expect(space.GUID).To(Equal("885735b5-aea4-4cf5-8e44-961af0e41920"))
				g.Expect(space.Name).To(Equal("my-space"))
				g.Expect(space.Relationships).To(HaveKeyWithValue("organization", model.Relationship{Data: model.RelationshipData{GUID: "e00705b9-7b42-4561-ae97-2520399d2133"}}))
			})
		})

		when("CF API is down", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.Close()
			})

			it("returns a meaningful error", func() {
				_, err := client.GetSpace("space-guid")

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("connection refused"))
			})
		})

		when("CF API returns a non-200 status code", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						gg.VerifyRequest("GET", "/v3/spaces/space-guid"),
						gg.RespondWith(418, ""),
					),
				)
			})

			it("returns a meaningful error", func() {
				_, err := client.GetSpace("space-guid")

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("failed to get space, received status: 418"))
			})
		})

		when("uaa client fails to fetch a token", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("", errors.New("uaa-fail"))
			})

			it("errors", func() {
				_, err := client.GetSpace("space-guid")

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("uaa-fail"))
			})
		})
	})
}

func TestClientGetDomain(t *testing.T) {
	spec.Run(t, "TestClientGetDomain", func(t *testing.T, when spec.G, it spec.S) {
		// TODO: remove `g` and `gg` after moving to ginkgo
		var (
			g               = NewWithT(t)
			gg              = ghttp.NewGHTTPWithGomega(g)
			client          *Client
			fakeCFAPIServer *ghttp.Server
		)

		it.Before(func() {
			fakeCFAPIServer = ghttp.NewServer()

			client = NewClient(fakeCFAPIServer.URL(), new(mocks.Rest), new(mocks.TokenFetcher))
		})

		it.After(func() {
			fakeCFAPIServer.Close()
			mock.AssertExpectationsForObjects(t, client.restClient, client.uaaClient)
		})

		when("CF API is operating normally", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						gg.VerifyRequest("GET", "/v3/domains/3a5d3d89-3f89-4f05-8188-8a2b298c79d5"),
						gg.RespondWith(200, `{
  "guid": "3a5d3d89-3f89-4f05-8188-8a2b298c79d5",
  "created_at": "2019-03-08T01:06:19Z",
  "updated_at": "2019-03-08T01:06:19Z",
  "name": "test-domain.com",
  "internal": false,
  "router_group": null,
  "supported_protocols": ["http"],
  "metadata": {
    "labels": { },
    "annotations": { }
  },
  "relationships": {
    "organization": {
      "data": { "guid": "3a3f3d89-3f89-4f05-8188-751b298c79d5" }
    },
    "shared_organizations": {
      "data": [
        {"guid": "404f3d89-3f89-6z72-8188-751b298d88d5"},
        {"guid": "416d3d89-3f89-8h67-2189-123b298d3592"}
      ]
    }
  },
  "links": {
    "self": {
      "href": "https://api.example.org/v3/domains/3a5d3d89-3f89-4f05-8188-8a2b298c79d5"
    },
    "organization": {
      "href": "https://api.example.org/v3/organizations/3a3f3d89-3f89-4f05-8188-751b298c79d5"
    },
    "route_reservations": {
      "href": "https://api.example.org/v3/domains/3a5d3d89-3f89-4f05-8188-8a2b298c79d5/route_reservations"
    },
    "shared_organizations": {
      "href": "https://api.example.org/v3/domains/3a5d3d89-3f89-4f05-8188-8a2b298c79d5/relationships/shared_organizations"
    }
  }
}`),
					),
				)
			})

			it("returns an expected domain object", func() {
				domain, err := client.GetDomain("3a5d3d89-3f89-4f05-8188-8a2b298c79d5")

				g.Expect(err).To(BeNil())
				g.Expect(domain).ToNot(BeNil())

				g.Expect(domain.GUID).To(Equal("3a5d3d89-3f89-4f05-8188-8a2b298c79d5"))
				g.Expect(domain.Name).To(Equal("test-domain.com"))
				g.Expect(domain.Internal).To(BeFalse())
			})
		})

		when("CF API is down", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.Close()
			})

			it("returns a meaningful error", func() {
				_, err := client.GetDomain("domain-guid")

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("connection refused"))
			})
		})

		when("CF API returns a non-200 status code", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("valid-token", nil)
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						gg.VerifyRequest("GET", "/v3/domains/domain-guid"),
						gg.RespondWith(418, ""),
					),
				)
			})

			it("returns a meaningful error", func() {
				_, err := client.GetDomain("domain-guid")

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("failed to get domain, received status: 418"))
			})
		})

		when("uaa client fails to fetch a token", func() {
			it.Before(func() {
				client.uaaClient.(*mocks.TokenFetcher).On("Fetch").Return("", errors.New("uaa-fail"))
			})

			it("errors", func() {
				_, err := client.GetDomain("domain-guid")

				g.Expect(err).ToNot(BeNil())
				g.Expect(err.Error()).To(ContainSubstring("uaa-fail"))
			})
		})
	})
}
