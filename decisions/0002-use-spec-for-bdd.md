# 1. Use spec for bdd in capi-k8s-watcher

Date: 2020-1-05 (roughly)

## Status

Draft

## Context

We wanted a test framework to provide some scaffolding for testing new k8s code.
Some folks on the team had explained Ginkgo's problems with test pollution and
concurrency to many-a-pair. @christarazi wanted to use tools with broader
acceptance in the k8s/go community. Testify vs Gomega also came into play.

## Decision

We will try Stephen Levine's ["spec" library](https://github.com/sclevine/spec)
We will try [testify](https://github.com/stretchr/testify) while we're at it.

## Consequences

We can run tests via "go test" and tests work better in IDEs like Goland.

Tests are serial by default, but we can control concurrency more granularly
using the same subtest mechanisms as go's default test suite.

By default, spec doesn't come plugged in with Gomega, so there's concern that
test output might be worse if we choose to use it in the future.

Testify is more widely used than Gomega and may be familiar to the broader go
community, but also asserts expectations all at once in an After
(mock.AssertExpectationsForObjects) in a way that might be risky on complicated
test cases.
