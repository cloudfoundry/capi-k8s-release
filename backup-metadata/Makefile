.PHONY: all artifact build test-unit test-binary test-e2e

all: build

artifact: cf-metadata.tgz

cf-metadata.tgz: bin/* config/*
	cd .. ; tar -cf backup-metadata/cf-metadata.tgz backup-metadata/bin/* backup-metadata/config/*

build: test-unit test-binary
	cd src && go build -o backup-metadata cmd/main.go

test-unit:
	cd src && ginkgo -r internal 

test-binary:
	cd src && ginkgo -r test/binary

test-e2e:
	cd src && ginkgo -r test/e2e

