# 4. Sync Routes from k8s to CC

Date: 2020-06-23

## Status

Proposal

## Context

### CF History

To maintain reliability in the event of temporary network partition, CF must ensure the user's desired state is consistent across components.

In the past, we've done this using "bulk sync" agents like: 
1. [nsync-bulker](https://github.com/cloudfoundry-attic/nsync/tree/master/cmd/nsync-bulker)
1. [ProcessesSync](https://github.com/cloudfoundry/cloud_controller_ng/blob/master/lib/cloud_controller/diego/processes_sync.rb)
1. [Copilot::Sync](https://github.com/cloudfoundry/cloud_controller_ng/blob/master/lib/cloud_controller/copilot/sync.rb)
1. [hm9000](https://github.com/cloudfoundry-attic/hm9000)

Each of these components share a similar goal: fetch desired state and ensure it is consistently represented everywhere where it needs to become actual state.
In practice, the modern versions of these bulk sync loops page through CCDB or the CF API and make requests to external components (diego, copilot, metac) to ensure that their internal states are consistent with values in CCDB.

### Kubernetes

Kubernetes is also implemented via a series of "bulk sync loops" that are more commonly known as controllers, but there are some key differences in how these convergence loops work.
1. k8s has a ["watch"](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes) API so clients clients can subscribe to incremental changes in desired and actual state
1. [k8s.io/client-go/cache](https://godoc.org/k8s.io/client-go/tools/cache) provides a mechanism for clients to build and maintain local caches of kubernetes state
1. [kubebuilder](https://book.kubebuilder.io/) Provides, amongst other things, an SDK for building reconciliation loops on top of watches and client side caching

### Problems and Solutions

To illustrate the necessity of these sync loops, it helps to think about how they improve resiliency in 2 cases of temporary network partition from the perspective of the CF API: 
1. Problem: Updates to CCDB succeed ✅, but updates to the Route CRD fail ❌
1. Problem: Updates to the Route CRD succeed ✅, but updates to CCDB fail ❌

To handle these 2 failure modes, several approaches can be taken. 

##### Mimic historical sync loops:
1. Problem: CCDB ✅ & Route CRD ❌
   1. Solution: Raise the error to the user, but allow the CCDB transaction to commit. A bulk sync loop will detect that the Route CRD is "missing" and create it later.
1. Problem: Route CRD ✅ & CCDB ❌
   1. Solution A: Raise the error to the user, and don't do the Route CRD write until CCDB changes are committed.
   1. Solution B: A bulk sync loop will detect that the Route CRD is "extraneous" and delete it later.

##### Forgo a bulk sync and attempt to bootstrap consistency with k8s off CCDB transactions:
1. Problem: CCDB ✅ & Route CRD ❌
   1. Solution: Raise the error to the user, abort the CCDB transaction.
1. Problem: Route CRD ✅ & CCDB ❌
   1. Solution: Roll back changes to the Route CRD... (but what if that still fails due to network partition?!)

##### Sync from k8s to CCDB
1. Problem: CCDB ✅ & Route CRD ❌
   1. Solution: Raise the error to the user, abort the CCDB transaction.
1. Problem: Route CRD ✅ & CCDB ❌
   1. Solution: Raise the error to the user, but leave the Route CRD. The "bulk sync loop" will not consider the route "reconciled" until it is correct in CCDB.

### Viable Long Term Implementations & Tradeoffs

##### CCDB->k8s Sync in Ruby
1. Mimics ProcessesSync for efficient access to CCDB
1. Run it in an independent process from clock_global to avoid BOSH entanglement
1. Unclear what scale limitations we might encounter if we regularly download
   all k8s Routes.

##### CF API->k8s Sync in Go
1. Probably a new, independent go program
1. Can take advantage of go-client cache & reflector to avoid repeatedly
   downloading all k8s state.
   1. What does it mean to "reconcile" a change to a single Route CRD if CCDB is authoritative?
1. Would need to page through CF API to build picture of CF state.

##### Sync from k8s->CCDB with kubebuilder controller and reloads
1. Probably a RoutesController akin to our existing kpack BuildsController
1. Can take advantage of go-client cache & reflector to avoid repeatedly
   downloading all k8s state.
   1. Can take advantage of existing kubebuilder patterns
   1. "Reconciling" a change to a single Route CRD means verifying with the CF
      API that it has been loaded from k8s.
1. Route Reconcile must `POST /v3/route/:guid/reload` endpoint on CF API
   1. These must only succeed if the given resource can be loaded from k8s
   1. Each "reconcile" call requires a successful call to this endpoint or else
      it'll be reqeued.
   1. Endpoint must short circuit if provided k8s resourceVersion matches one
      stored in CCDB. This avoids endless loops and avoids unnecessary k8s
      traffic.
1. In normal operation, CF API must be careful not to commit changes to CCDB
   that failed to be applied to k8s.
   1. In practice, this means k8s Routes operations must occur inside CCDB
      transactions.
1. In failure modes, restarting RoutesController would resync existing Routes.
1. Hard to handle Routes that are missing in k8s but still present in CCDB.
   1. In theory this should never happen as the k8s delete would need to be
      reconciled, but in practice it could happen due to bugs
   1. Might require us to periodically `POST /v3/routes/:guid/reload` every route
      in CCDB, anyways, or just have operators do that in the rare case the
      route counts are out-of-sync.
   1. Maybe PATCHes cause deletes with 409 Gone?
1. Allows platform operators to add routes as part of their cf-for-k8s
   configuration that could be viewable via the CF API.

## Decision

We ended up taking a bit of a hybrid approach between "CF API->k8s Sync in Go" and "Sync from k8s->CCDB with kubebuilder controller and reloads".
In the discussion of this ADR it became apparent that we could use kubebuilder to build a bulk sync loop that reconciles all routes from the CF API.

We added a new CRD (PeriodicSync) and controller to cf-api-controllers. Using the CF API as a source of truth, the PeriodicSyncReconciler runs in the existing cf-ai-controller-manager and reconciles all the /v3/routes periodically.

## Consequences

### Pros
1. Correct behavior for "extra" routes in k8s.
    1. This case was hard to solve with any other proposal
1. Minimal new Ruby code, minimal unnecessary k8s API load
1. No custom logic for managing periodicity, and no reliance on CC's existing clock mechanism thanks to kubebuilder
1. Doesn't prevent us from optimizing using per-route reconciliation later
    1. When the route-doesn't-exist in k8s, this might be a little wonky, but is likely still possible
1. PeriodicSync CRD has "type" field so we can hopefully re-use some of this work for other resources like Droplets or LRPs.
1. This loop would also be necessary syncing k8s->CCDB.
    1. For any sync direction, you can't detect an extra CCDB route without a loop like this.

### Cons
1. Similar to existing sync loops, care must be taken to prevent the loop from exiting early on errors
    1. It does do this now, but it's not always easy
1. No CCDB access in cf-api-controllers
    1. Serializing/deserializing /v3/routes frequently
    1. We punted on pagination, limiting us to 5000 routes. Handling pagination in the future might be challenging, especially if it can cause us to miss routes in CCDB.
