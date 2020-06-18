# 1. Record architecture decisions

Date: 2020-06-17

## Status

In Review

## Context

Currently we store capi-kpack-watcher directly in capi-k8s-release. This has
caused some complexity when bumping image versions in our concourse pipeline.
We're also planning on reworking the watcher to use kubebuilder, and so now is a
good time to change any source code organization decisions.

Proposed locations for capi-kpack-watcher source code:

1. `cf-api-kpack-watcher` a repo just for this particular component

   1. Pros:

      1. Very clear what goes where

      1. It'd be a standard kubebuilder go repo

      1. Fits our current pattern of go service repositories (1 repo for each service)

   1. Cons:

      1. Many more repos to manage to keep this pattern consistent in the
       near-future with routes, droplets, etc

1. `cf-api-controllers`/`cf-controllers` which would contain other future controllers we create for CF API resources

   1. Pros:

      1. It'd be a standard kubebuilder go repo with multiple controllers

      1. Only one more repo to manage to keep this pattern consistent

   1. Cons:

      1. Possibly unclear in the future whether controllers without CF API dependency should go here.

1. `cloud_controller_ng` mono-repo

   1. Pros:

      1. 0 more repos to manage to keep this pattern consistent

      1. No way for API to get out-of-sync with controllers - they all come from
       the same source code.

   1. Cons:

      1. 1 repo with Ruby and Go code inside can be annoying to grep.

      1. Nonstandard kubebuilder layout

1. leave it in `capi-k8s-release/src/capi-kpack-watcher`

   1. Pros:

      1. 0 more repos to manage to keep this pattern consistent

   1. Cons:

      1. Bumping capi-k8s-release's built image reference with a SHA that points
       at itself is confusing and difficult to manage within concourse.

      1. Nonstandard kubebuilder layout

## Decision




## Consequences

