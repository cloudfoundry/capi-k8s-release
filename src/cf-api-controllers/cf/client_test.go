package cf_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/onsi/gomega/ghttp"

	. "code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/cffakes"

	. "github.com/onsi/ginkgo"

	"code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers/cf/model"

	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		client       *Client
		restClient   *cffakes.FakeRest
		tokenFetcher *cffakes.FakeTokenFetcher
	)

	BeforeEach(func() {
		tokenFetcher = new(cffakes.FakeTokenFetcher)
		tokenFetcher.FetchReturns("valid-token", nil)

		restClient = new(cffakes.FakeRest)
		client = NewClient("http://capi.host", restClient, tokenFetcher)
	})

	Describe("UpdateBuild", func() {
		const (
			guid = "guid"
		)

		var (
			build model.Build
		)

		When("successfully updates", func() {
			BeforeEach(func() {
				build = model.Build{State: "SUCCESS"}
				restClient.PatchReturns(&http.Response{StatusCode: 200}, nil)
			})

			It("fetches a token and updates CF API server", func() {
				Expect(client.UpdateBuild(guid, build)).To(Succeed())

				Expect(tokenFetcher.FetchCallCount()).To(Equal(1))

				raw, err := json.Marshal(build)
				Expect(err).To(BeNil())

				Expect(restClient.PatchCallCount()).To(Equal(1))

				actualURL, actualAuthToken, actualBodyIO := restClient.PatchArgsForCall(0)
				Expect(actualURL).To(Equal("http://capi.host/v3/builds/guid"))
				Expect(actualAuthToken).To(Equal("valid-token"))

				actualBody, err := ioutil.ReadAll(actualBodyIO)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualBody).To(Equal(raw))
			})
		})

		When("uaa client fails to fetch a token", func() {
			BeforeEach(func() {
				tokenFetcher.FetchReturns("", errors.New("fail"))
			})

			It("errors", func() {
				Expect(client.UpdateBuild(guid, build)).ToNot(Succeed())
			})
		})

		When("CF API server client fails to Patch", func() {
			BeforeEach(func() {
				restClient.PatchReturns(nil, errors.New("fail"))
			})

			It("errors", func() {
				Expect(client.UpdateBuild(guid, build)).ToNot(Succeed())
			})
		})

		When("a non-400+ status code is received", func() {
			BeforeEach(func() {
				restClient.PatchReturns(&http.Response{StatusCode: 500}, nil)
			})

			It("errors", func() {
				Expect(client.UpdateBuild(guid, build)).ToNot(Succeed())
			})
		})
	})

	Describe("UpdateDroplet", func() {
		const (
			guid = "guid"
		)
		var (
			droplet model.Droplet
		)

		When("successfully updates", func() {
			BeforeEach(func() {
				droplet = model.Droplet{Image: "updated-image-reference"}
				restClient.PatchReturns(&http.Response{StatusCode: 200}, nil)
			})

			It("fetches a token and updates CF API server", func() {
				Expect(client.UpdateDroplet(guid, droplet)).To(Succeed())

				Expect(tokenFetcher.FetchCallCount()).To(Equal(1))

				raw, err := json.Marshal(droplet)
				Expect(err).To(BeNil())

				Expect(restClient.PatchCallCount()).To(Equal(1))

				actualURL, actualAuthToken, actualBodyIO := restClient.PatchArgsForCall(0)
				Expect(actualURL).To(Equal("http://capi.host/v3/droplets/guid"))
				Expect(actualAuthToken).To(Equal("valid-token"))

				actualBody, err := ioutil.ReadAll(actualBodyIO)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualBody).To(Equal(raw))
			})
		})

		When("uaa client fails to fetch a token", func() {
			BeforeEach(func() {
				tokenFetcher.FetchReturns("", errors.New("fail"))
			})

			It("errors", func() {
				Expect(client.UpdateDroplet(guid, droplet)).ToNot(Succeed())
			})
		})

		When("CF API server client fails to Patch", func() {
			BeforeEach(func() {
				restClient.PatchReturns(nil, errors.New("fail"))
			})

			It("errors", func() {
				Expect(client.UpdateDroplet(guid, droplet)).ToNot(Succeed())
			})
		})

		When("a non-400+ status code is received", func() {
			BeforeEach(func() {
				restClient.PatchReturns(&http.Response{StatusCode: 500}, nil)
			})

			It("errors", func() {
				Expect(client.UpdateDroplet(guid, droplet)).ToNot(Succeed())
			})
		})
	})

	Describe("ListRoutes", func() {
		var (
			fakeCFAPIServer *ghttp.Server
		)

		BeforeEach(func() {
			fakeCFAPIServer = ghttp.NewServer()

			client = NewClient(fakeCFAPIServer.URL(), restClient, tokenFetcher)
		})

		AfterEach(func() {
			fakeCFAPIServer.Close()
		})

		When("CF API is operating normally", func() {
			BeforeEach(func() {
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v3/routes"),
						ghttp.RespondWith(200, `{
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
	 ],
"included": {
    "spaces": [
      {
        "guid": "885a8cb3-c07b-4856-b448-eeb10bf36236",
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
            "href": "https://api.example.org/v3/spaces/885a8cb3-c07b-4856-b448-eeb10bf36236"
          },
          "features": {
            "href": "https://api.example.org/v3/spaces/885a8cb3-c07b-4856-b448-eeb10bf36236/features"
          },
          "organization": {
            "href": "https://api.example.org/v3/organizations/e00705b9-7b42-4561-ae97-2520399d2133"
          },
          "apply_manifest": {
            "href": "https://api.example.org/v3/spaces/885a8cb3-c07b-4856-b448-eeb10bf36236/actions/apply_manifest",
            "method": "POST"
          }
        },
        "metadata": {
          "labels": {},
          "annotations": {}
        }
      }
    ],
    "domains": [
      {
        "guid": "0b5f3633-194c-42d2-9408-972366617e0e",
        "created_at": "2019-03-08T01:06:19Z",
        "updated_at": "2019-03-08T01:06:19Z",
        "name": "test-domain.com",
        "internal": false,
        "router_group": null,
        "supported_protocols": [
          "http"
        ],
        "metadata": {
          "labels": {},
          "annotations": {}
        },
        "relationships": {
          "organization": {
            "data": {
              "guid": "e00705b9-7b42-4561-ae97-2520399d2133"
            }
          },
          "shared_organizations": {
            "data": [
              {
                "guid": "404f3d89-3f89-6z72-8188-751b298d88d5"
              },
              {
                "guid": "416d3d89-3f89-8h67-2189-123b298d3592"
              }
            ]
          }
        },
        "links": {
          "self": {
            "href": "https://api.example.org/v3/domains/0b5f3633-194c-42d2-9408-972366617e0e"
          },
          "organization": {
            "href": "https://api.example.org/v3/organizations/e00705b9-7b42-4561-ae97-2520399d2133"
          },
          "route_reservations": {
            "href": "https://api.example.org/v3/domains/0b5f3633-194c-42d2-9408-972366617e0e/route_reservations"
          },
          "shared_organizations": {
            "href": "https://api.example.org/v3/domains/0b5f3633-194c-42d2-9408-972366617e0e/relationships/shared_organizations"
          }
        }
      }
    ]
}
	}`),
					),
				)
			})

			It("returns a list of routes", func() {
				routeList, err := client.ListRoutes()
				routes := routeList.Resources

				Expect(err).To(BeNil())
				Expect(routes).ToNot(BeEmpty())
				Expect(routes).To(HaveLen(1))

				Expect(routes[0].GUID).To(Equal("cbad697f-cac1-48f4-9017-ac08f39dfb31"))
				Expect(routes[0].Host).To(Equal("a-hostname"))
				Expect(routes[0].Path).To(Equal("/some_path"))
				Expect(routes[0].URL).To(Equal("a-hostname.a-domain.com/some_path"))

				Expect(routes[0].Destinations).To(HaveLen(2))
				Expect(routes[0].Destinations[0].GUID).To(Equal("385bf117-17f5-4689-8c5c-08c6cc821fed"))
				Expect(routes[0].Destinations[0].Port).To(Equal(8080))
				Expect(routes[0].Destinations[0].Weight).To(BeNil())
				Expect(routes[0].Destinations[0].App.GUID).To(Equal("0a6636b5-7fc4-44d8-8752-0db3e40b35a5"))
				Expect(routes[0].Destinations[0].App.Process.Type).To(Equal("web"))

				Expect(routes[0].Relationships).To(HaveKeyWithValue("space", model.Relationship{Data: model.RelationshipData{GUID: "885a8cb3-c07b-4856-b448-eeb10bf36236"}}))
				Expect(routes[0].Relationships).To(HaveKeyWithValue("domain", model.Relationship{Data: model.RelationshipData{GUID: "0b5f3633-194c-42d2-9408-972366617e0e"}}))

				space := routeList.Included.Spaces[0]
				Expect(space.Name).To(Equal("my-space"))
				Expect(space.GUID).To(Equal("885a8cb3-c07b-4856-b448-eeb10bf36236"))
				Expect(space.Relationships).To(HaveKeyWithValue("organization", model.Relationship{Data: model.RelationshipData{GUID: "e00705b9-7b42-4561-ae97-2520399d2133"}}))

				domain := routeList.Included.Domains[0]
				Expect(domain.GUID).To(Equal("0b5f3633-194c-42d2-9408-972366617e0e"))
				Expect(domain.Name).To(Equal("test-domain.com"))
				Expect(domain.Internal).To(BeFalse())
			})
		})

		When("CF API is down", func() {
			BeforeEach(func() {
				fakeCFAPIServer.Close()
			})

			It("returns a meaningful error", func() {
				_, err := client.ListRoutes()

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("connection refused"))
			})
		})

		When("CF API returns a non-200 status code", func() {
			BeforeEach(func() {
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v3/routes"),
						ghttp.RespondWith(418, ""),
					),
				)
			})

			It("returns a meaningful error", func() {
				_, err := client.ListRoutes()

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to list routes, received status: 418"))
			})
		})

		When("CF API returns an unexpected JSON response", func() {
			BeforeEach(func() {
				fakeCFAPIServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v3/routes"),
						ghttp.RespondWith(200, `{`),
					),
				)
			})

			It("returns a meaningful error", func() {
				_, err := client.ListRoutes()

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to deserialize response from CF API"))
			})
		})

		When("uaa client fails to fetch a token", func() {
			BeforeEach(func() {
				tokenFetcher.FetchReturns("", errors.New("uaa-fail"))
			})

			It("errors", func() {
				_, err := client.ListRoutes()

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("uaa-fail"))
			})
		})
	})
})
