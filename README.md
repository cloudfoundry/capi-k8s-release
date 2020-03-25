### Disclaimer
This repo is very new, experimental, and not yet continuously integrated or tested to the degree we'd like.

### Pre-requisites

1. Clone the `cf-for-k8s` repository: https://github.com/cloudfoundry/cf-for-k8s/
1. Install any prerequisites of `cf-for-k8s`: https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/deploy.md#prerequisites

### Configuring pushes of buildpack apps

`capi-k8s-release` currently uploads app source code to a blobstore, but then hands that off to `kpack` to build app images that are then placed in a registry.  In order for this to work, you must configure the following values:

```yaml
kpack:
  registry:
    hostname: # the hostname of the registry, used for authentication
    repository: # the destination of the build app images within the registry
    username: # basic auth registry username
    password: # basic auth registry password
```

dockerhub example:
```yaml
kpack:
  registry:
    hostname: https://index.docker.io/v1/
    repository: cloudfoundry/capi
    username: <username>
    password: <password>
```

gcr example:
```yaml
kpack:
  registry:
    hostname: gcr.io
    repository: gcr.io/cloudfoundry/capi
    username: <username>
    password: <password>
```


### Rolling out changes to CAPI

1. `./scripts/rollout.sh` will take any local changes to `capi-k8s-release`, apply them to a local `cf-for-k8s` directory, and then deploy `cf-for-k8s`
  - Local `cf-for-k8s` directory can be overriden by setting `CF_FOR_K8s_DIR`
