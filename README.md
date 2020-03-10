### Disclaimer
This repo is very new, experimental, and not yet continuously integrated or tested to the degree we'd like.

### Pre-requisites

1. Clone the `cf-for-k8s` repository: https://github.com/cloudfoundry/cf-for-k8s/
1. Install any prerequisites of `cf-for-k8s`: https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/deploy.md#prerequisites

### Rolling out changes to CAPI

1. `./scripts/rollout.sh` will take any local changes to `capi-k8s-release`, apply them to a local `cf-for-k8s` directory, and then deploy `cf-for-k8s`
  - Local `cf-for-k8s` directory can be overriden by setting `CF_FOR_K8s_DIR`
