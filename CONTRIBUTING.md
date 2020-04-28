# Prerequisites
1. Clone the `cf-for-k8s` repository: https://github.com/cloudfoundry/cf-for-k8s/
1. Install any prerequisites of `cf-for-k8s`: https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/deploy.md#prerequisites
1. Get [`kbld`](https://k14s.io/#install)

# Deploying Local Changes

## Deploying New `cloud_controller_ng` and `capi-k8s-release` Changes
If you want to deploy changes of the `cloud_controller_ng` source along with changes of `capi-k8s-release` templates to your `cf-for-k8s` deployment, follow these steps:

1. Ensure you have a Docker registry (e.g. Docker Hub, GCR) that you can push to and you are logged into it via `docker login`
1. Run the following:
    ```
    IMAGE_DESTINATION=<docker_registry_url>/<desired_repo_name> ./scripts/build-and-rollout.sh <path_to_cf_install_values_file>
    ```

## Deploying Only New `capi-k8s-release` Changes
If you want to deploy changes of just the `capi-k8s-release` templates to your `cf-for-k8s` deployment, follow these steps:

1. Ensure you have a Docker registry (e.g. Docker Hub, GCR) that you can push to and you are logged into it via `docker login`
1. Run the following:
    ```
    ./scripts/rollout.sh <path_to_cf_install_values_file>
    ```

# Testing Changes
- Run `cf-for-k8s` smoke tests: https://github.com/cloudfoundry/cf-for-k8s/blob/master/docs/contributing.md#running-smoke-tests
- Run CAPI BARAS: https://github.com/cloudfoundry/capi-bara-tests/blob/master/README.md
  - All of the necessary configuration information to run CAPI BARAS should be inside of the values file you used to deploy `cf-for-k8s` (e.g. `admin_password` and `apps_domain`)
  - Also you **must** include the following configuration for CAPI BARAS against a `cf-for-k8s` environment: `"include_kpack": true`

# Contributing Changes
Ensure you have tested your changes and submit a PR to this repository against the `master` branch.

Feel free to comment on the PR itself or reach out to us on [#capi in Cloud Foundry slack](https://cloudfoundry.slack.com/archives/C07C04W4Q) and ping `@interrupt` for assistance

