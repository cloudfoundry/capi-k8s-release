# Installation Steps

## Prequisite
- You have a kubernetes cluster
- You have backup-metadata tarball downloaded (`gs://tas-cf-metadata-test-artifacts`)

## Installation instruction
- Extract the tarball:
`tar xvf backup-metadata.*.tgz && cd backup-metadata`
- Create a folder `env-config`
- Create a yaml file `values.yml` in `env-config` folder with contents:
```
#@data/values
---
namespace: ""

cf:
    api: ""
    admin_username: ""
    admin_password: ""

registry:
  hostname: ""
  username: ""
  password: ""

```
- Fill in values in `env-config/values.yml`
- execute installation script with values folder
`./bin/install.sh env-config`