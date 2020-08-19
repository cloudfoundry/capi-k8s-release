# package-image-uploader

A web server that converts package zip files to single-layer OCI images and uploads them to a registry.

## Usage
To start the server locally, run `go run code.cloudfoundry.org/capi-k8s-release/src/package-image-uploader/cmd/server`

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
