#!/usr/bin/env bash

 kubectl exec \
   -n cf-system \
   -it deployments/cf-api-clock \
   -- /cloud_controller_ng/bin/console
