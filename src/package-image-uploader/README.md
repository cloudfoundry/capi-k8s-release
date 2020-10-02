# package-image-uploader
A minimal server for interacting with OCI image registries.
Designed to be colocated with the `cf-api-server` and `cf-api-workers`.

## Usage
### Prerequisites
The following environment variables must be configured:

* `REGISTRY_USERNAME`: Container registry username (e.g. DockerHub username or `_json_key` for GCR<sup>1</sup>)
* `REGISTRY_PASSWORD`: Container registry credentials (e.g. DockerHub password or GCR service account json)
* `PORT`: Port the server will listen on. Default: `8080`

<sup>1</sup> For more information on GCR authentication [check out these docs](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key).

To start the server locally, run:
 ```
go run code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/cmd/server
 ```

### POST /packages
Converts package zip files to single-layer OCI images and uploads them to the specified registry.
**Note:** `package_zip_path` must refer to an accessible local file path.

Request body:
```
{
  "package_zip_path": "/path/to/package.zip",
  "package_guid": "a-package-guid",
  "registry_base_path": "docker.io/cfcapidocker"
}
```

Response code: `200`

Response body:
```
{
  "hash": {
    "algorithm": "sha256",
    "hex": "a03c91dbeb4e7cf53862c8c96624d2922448276162f3485a03e7c95bd82937ef"
  }
}
```

### DELETE /images
Deletes an image from the registry. The image reference should include a tag or digest. Defaults to the `latest` tag.
When an image is deleted by tag the endpoint will attempt to delete the manifests for both the tag and the digest.
**Note:** Many registries delete images asynchronously, so the image may not be deleted immediately.
**Note:** Some registries (e.g. GCR) require deleting an image's tag manifests before its digest manifest can be deleted.

Request body:
```
{
  "image_reference": "docker.io/cfcapidocker/some-image-name:some-tag",
}
```

Response code: `202`

Response body:
```
{
  "image_reference": "docker.io/cfcapidocker/some-image-name:some-tag",
}
```
