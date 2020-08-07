#!/usr/bin/env bash

 kubectl -n cf-system exec -it deployments/cf-api-server -- /cloud_controller_ng/bin/console /config/cloud_controller_ng.yml
