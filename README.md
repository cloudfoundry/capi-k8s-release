### Pre-requisites

1. `minikube` (`brew install minikube`)
1. `helm` (`brew install kubernetes-helm`) We require helm 3.0.0+ 
1. `minikube addons enable registry`
1. Add the `minikube`'s IP (`minikube ip`) to the list of insecure docker registries of your
   local docker daemon (Either using the Docker Desktop or editing
   `~/.docker/daemon.json`):
```
{
  "insecure-registries" : ["minikubeIP:5000"]
}
```
Then, restart Docker. (Docker Desktop icon > Restart)
1. To ensure cloud controller submodule is up-to-date run `git submodule update --init`


### Installing dependencies and CAPI

CAPI requires a database and blobstore.  We chose to use Postgres and Minio for
those dependencies, respectively.  Both have stable `helm` charts, so that is
the approach we use to install them.


1. `minikube start` to make sure `minikube` is up and running
1. `./scripts/deploy.sh` to deploy the dependencies and CAPI


### Rolling out changes to CAPI

1. `./scripts/build-and-rollout-capi.sh` will take the `cloud_controller_ng` code in
   the `src/cloud_controller_ng` submodule, build a docker image with it, and
   roll the new image out to the `minikube` cluster.

### Known Issues

1. If you see an issue with helm (eg. tiller not found) then update to the latest helm version (`brew upgrade helm`)

