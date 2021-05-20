# 5. Use Exec / Curl in Liveness Probe for Registry-Buddy

Date: 2021-05-20

## Status

Accepted

## Context

In [#178030273](https://www.pivotaltracker.com/story/show/178030273) we added a liveness probe to the `registry-buddy`
containers so that they will be restarted when the container is unhealthy. The `registry-buddy` is a simple http server
implemented in Go that is configured with permissions to push and delete images from a container registry.

It does not perform any authentication or authorization checks of its callers and is only protected by virtue of it being
a sidecar container on trusted workloads (`cf-api-server` and `cf-api-worker`). Given that it makes no auth(n/z) checks
we have it binding only to the loopback address (e.g. localhost). We cannot use the regular [`http` liveness probe](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-http-request)
to health check it since the `kubelet` does not have access to `localhost` within the Pod's network namespace.

To enable the `kubelet` to reach the `registry-buddy` we would need to listen on a reachable address instead of `localhost`.
This may expose it to the rest of the Pod network, though, and open up the registry to access by non-trusted `Pods`.

## Decision

Instead of adding additional auth(n/z) to `registry-buddy` and having it listen on an externally reachable address, we chose
to instead use an `exec` liveness probe to execute a `curl` within the container itself.

## Consequences

`registry-buddy` can continue listening on `localhost` and we can defer the work to add additional auth(n/z).
