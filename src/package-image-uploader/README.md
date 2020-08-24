# package-image-uploader

A web server that converts package zip files to single-layer OCI images and uploads them to a registry.

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
Request body
```
{
  "package_zip_path": "/path/to/package.zip",
  "package_guid": "a-package-guid",
  "registry_base_path": "docker.io/cfcapidocker"
}
```

Response body
```
{
  "hash": {
    "algorithm": "sha256",
    "hex": "a03c91dbeb4e7cf53862c8c96624d2922448276162f3485a03e7c95bd82937ef"
  }
}
```
