# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"
# if kubebuilder isn't present in /usr/local/kubebuilder , install this version
KUBEBUILDER_VERSION ?= 2.3.1
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
test: generate fmt vet /usr/local/kubebuilder/
	go test -v ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager -v main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./main.go

# Install CRDs into a cluster
install:
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall:
	kustomize build config/crd | kubectl delete -f -

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./..." output:crd:artifacts:config=config/crd/bases
	cp config/crd/bases/apps.cloudfoundry.org_periodicsyncs.yaml ../../config/periodic-sync-crd.yml

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

/usr/local/kubebuilder/:
	# download kubebuilder and extract it to tmp
	curl -L https://go.kubebuilder.io/dl/${KUBEBUILDER_VERSION}/${GOOS}/${GOARCH} | tar -xz -C /tmp/

	# move to a long-term location and put it on your path
	sudo mv /tmp/kubebuilder_${KUBEBUILDER_VERSION}_${GOOS}_${GOARCH} /usr/local/kubebuilder

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
