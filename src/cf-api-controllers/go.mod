module code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers

go 1.14

require (
	github.com/buildpacks/lifecycle v0.8.1
	github.com/cloudfoundry-community/go-uaa v0.3.1
	github.com/go-logr/logr v0.1.0
	github.com/google/go-containerregistry v0.1.1
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pivotal/kpack v0.0.10-0.20200715191345-68925eaca94b
	github.com/sclevine/spec v1.4.0
	github.com/stretchr/testify v1.6.1
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	k8s.io/api v0.17.5
	k8s.io/apimachinery v0.17.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	sigs.k8s.io/controller-runtime v0.5.0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.5
