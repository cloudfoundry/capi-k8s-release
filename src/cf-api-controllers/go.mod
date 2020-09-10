module code.cloudfoundry.org/capi-k8s-release/src/cf-api-controllers

go 1.14

require (
	code.cloudfoundry.org/cf-k8s-networking/routecontroller v0.0.0-20200903181459-ab1b32f57e7c
	github.com/buildpacks/lifecycle v0.9.1
	github.com/cloudfoundry-community/go-uaa v0.3.1
	github.com/go-logr/logr v0.1.0
	github.com/google/go-containerregistry v0.1.1
	github.com/matt-royal/biloba v0.2.1
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/pivotal/kpack v0.1.2
	golang.org/x/net v0.0.0-20200813134508-3edf25e44fcc // indirect
	golang.org/x/sys v0.0.0-20200824131525-c12d262b63d8 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	sigs.k8s.io/controller-runtime v0.5.0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.5
