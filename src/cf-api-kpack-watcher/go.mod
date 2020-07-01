module code.cloudfoundry.org/capi-k8s-release/src/cf-api-kpack-watcher

go 1.14

require (
	code.cloudfoundry.org/clock v1.0.0
	code.cloudfoundry.org/lager v2.0.0+incompatible
	code.cloudfoundry.org/trace-logger v0.0.0-20170119230301-107ef08a939d // indirect
	code.cloudfoundry.org/uaa-go-client v0.0.0-20200427231439-19a7eb57a1dc
	github.com/buildpacks/lifecycle v0.8.0
	github.com/go-logr/logr v0.1.0
	github.com/google/go-containerregistry v0.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pivotal/kpack v0.0.9
	github.com/sclevine/spec v1.4.0
	github.com/stretchr/testify v1.6.0
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	k8s.io/api v0.17.5
	k8s.io/apimachinery v0.17.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	sigs.k8s.io/controller-runtime v0.5.0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.5
