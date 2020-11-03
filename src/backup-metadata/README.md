# backup-metadata-generator
A utility that scrapes Cloud Foundry APIs to provide metadata about a CF installation.
It is designed to be invoked as a [Velero hook](https://velero.io/docs/v1.5/backup-hooks/).

For more information, see [this doc](https://docs.google.com/document/d/1aR6_v0wTrSWpH9G2XqHcUTsdCNzP82ZsiWUxiN-4Zno/).

## Running Tests

### Unit/Binary Tests
```
go test -v ./internal/... ./test/binary/...
```

### End-to-End (e2e) Integration Tests
**Prerequisites:**

1. A Kubernetes cluster with [Velero installed](https://velero.io/docs/v1.5/basic-install/#install-and-configure-the-server-components). See our [CI script](https://github.com/cloudfoundry/capi-ci/blob/master/ci/backup-metadata/install-velero.sh) for a GCP example.
   For GCP you will also need to provide a service account with the following permissions:

   ```
   compute.disks.create
   compute.disks.createSnapshot
   compute.disks.get
   compute.snapshots.create
   compute.snapshots.delete
   compute.snapshots.get
   compute.snapshots.useReadOnly
   storage.objects.create
   storage.objects.get
   storage.objects.delete
   storage.objects.list
   compute.zones.get
   storage.buckets.get
   ```

   For other IaaSes refer to the Velero docs.
2. The [`velero` CLI installed locally](https://velero.io/docs/v1.5/basic-install/#install-the-cli)
3. `kubectl` targeted to this environment
4. GNU grep as `grep` on your PATH (`brew install grep`)

```
go test -v ./internal/e2e
```

## Deployment
The following environment variables are required:

* `CF_API_HOST` - URL for the CF API
* `CF_CLIENT` - Name of a UAA Client with `cloud_controller.read` and `cloud_controller.read_only_admin` authorities
* `CF_CLIENT_SECRET` - Secret for authenticating the client

You will also need to provide the following [Velero hook](https://velero.io/docs/v1.5/backup-hooks/#backup-hooks) annotations on the `Pod` (substitute `CONTAINER_NAME` with the name of the container):

```
pre.hook.backup.velero.io/container: CONTAINER_NAME
pre.hook.backup.velero.io/command: '["/cnb/process/generate-metadata"]'
```

In `cf-for-k8s` the `backup-metadata` container is [colocated on the `cf-api-controllers` Deployment](https://github.com/cloudfoundry/capi-k8s-release/blob/master/templates/controllers_deployment.yml).
