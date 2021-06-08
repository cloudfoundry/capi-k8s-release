# cf-api

## Purpose
This pipeline runs the CAPI K8s component tests and builds the associated
images. It also runs various integration tests (CATs, capi-baras, smoke tests) on cf-for-k8s.

## Groups
- capi-k8s-release: runs unit and integration tests
- ship-it: creates a new Github release

## Validation
There are several jobs in this pipeline that run unit and integration tests to validate the components in capi-k8s-release. The registry-buddy and CF API controller components run go tests while the `cc-tests` job runs `bundle exec rake` against both MySQL and PostgresQL DBs. Once these tests pass, the images are built and deployed along cf-for-k8s develop branch on one of the pooled environments. A subset of [CATs](http://github.com/cloudfoundry/cf-acceptance-tests), [capi-baras](https://github.com/cloudfoundry/capi-bara-tests), and smoke tests are then run against the deployed CF for K8s.


## Image building
This pipeline builds images via [`kbld`](https://carvel.dev/kbld/) which uses [`pack`](https://github.com/buildpacks/pack) or `docker build` (for nginx) to build the container images. OCI `source` and `revision` labels are added to the images as part of this process.
See the  [OCI Annotations repo](https://github.com/opencontainers/image-spec/blob/master/annotations.md) for more details regarding these labels.

For more information, check out the `build/build.sh` script to see how we generate the kbld config and build the images.

## Release
Once all the validation tests have passed, `the k8s-ci-passed` job updates the image reference templates with the newly build images.

To release a new capi-k8s-release, you must manually trigger the `ship-it-k8s` job. This will grab the latest github release, bumps the minor version, and releases a new minor version of capi-k8s-release.
