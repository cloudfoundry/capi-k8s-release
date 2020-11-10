.PHONY: all artifact build test-unit test-binary test-e2e

all: build

artifact: cf-metadata.tgz

cf-metadata.tgz: bin/* config/*
	cd .. ; tar -cf backup-metadata-generator/cf-metadata.tgz backup-metadata-generator/bin/* backup-metadata-generator/config/*

build: test-unit test-binary
	go build -o backup-metadata-generator main.go

test-unit:
	ginkgo -r internal 

test-binary:
	ginkgo -r test/binary

test-e2e:
	ginkgo -r test/e2e

