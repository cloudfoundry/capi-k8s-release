#!/usr/bin/env bash

CLUSTER_NAME=${CLUSTER_NAME:-$1}
SYS_DOMAIN=${SYS_DOMAIN:-$CLUSTER_NAME.k8s.capi.land}
APPS_DOMAIN=${APPS_DOMAIN:-apps.$CLUSTER_NAME.k8s.capi.land}
: "${SHARED_DNS_ZONE_NAME:="kubenetes-clusters"}"
: "${GKE_GCP_PROJECT:="cf-capi-arya"}"
: "${DNS_GCP_PROJECT:="cff-capi-dns"}"

echo "Discovering Istio Gateway LB IP"
external_static_ip=""
while [ -z $external_static_ip ]; do
    sleep 1
    external_static_ip=$(kubectl get services/istio-ingressgateway -n istio-system --output="jsonpath={.status.loadBalancer.ingress[0].ip}")
done

echo "Configuring DNS (${DNS_GCP_PROJECT} - ${SHARED_DNS_ZONE_NAME}) for external IP: ${external_static_ip}"
gcloud dns record-sets transaction start --project ${DNS_GCP_PROJECT} --zone="${SHARED_DNS_ZONE_NAME}"

# DNS for system domain
gcp_records_json="$( gcloud dns record-sets list --project ${DNS_GCP_PROJECT} --zone "${SHARED_DNS_ZONE_NAME}" --name "*.${SYS_DOMAIN}" --format=json )"
record_count="$( echo "${gcp_records_json}" | jq 'length' )"
if [ "${record_count}" != "0" ]; then
  existing_record_ip="$( echo "${gcp_records_json}" | jq -r '.[0].rrdatas | join(" ")' )"
  gcloud dns record-sets transaction remove --project ${DNS_GCP_PROJECT} --name "*.${SYS_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=60 "${existing_record_ip}" --verbosity=debug
fi
gcloud dns record-sets transaction add --project ${DNS_GCP_PROJECT} --name "*.${SYS_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=60 "${external_static_ip}" --verbosity=debug

# DNS for apps domain
gcp_records_json="$( gcloud dns record-sets list --project ${DNS_GCP_PROJECT} --zone "${SHARED_DNS_ZONE_NAME}" --name "*.${APPS_DOMAIN}" --format=json )"
record_count="$( echo "${gcp_records_json}" | jq 'length' )"
if [ "${record_count}" != "0" ]; then
  existing_record_ip="$( echo "${gcp_records_json}" | jq -r '.[0].rrdatas | join(" ")' )"
  gcloud dns record-sets transaction remove --project ${DNS_GCP_PROJECT} --name "*.${APPS_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=60 "${existing_record_ip}" --verbosity=debug
fi
gcloud dns record-sets transaction add --project ${DNS_GCP_PROJECT} --name "*.${APPS_DOMAIN}" --type=A --zone="${SHARED_DNS_ZONE_NAME}" --ttl=60 "${external_static_ip}" --verbosity=debug

echo "Contents of transaction.yaml:"
cat transaction.yaml
gcloud dns record-sets transaction execute --project ${DNS_GCP_PROJECT} --zone="${SHARED_DNS_ZONE_NAME}" --verbosity=debug

resolved_ip=''
set +o pipefail
sleep_time=5
while [ "$resolved_ip" != "$external_static_ip" ]; do
  echo "Waiting $sleep_time seconds for DNS to propagate..."
  sleep $sleep_time
  resolved_ip=$(nslookup "api.${SYS_DOMAIN}" | (grep ${external_static_ip} || true) | cut -d ' ' -f2)
  echo "Resolved IP: ${resolved_ip}, Actual IP: ${external_static_ip}"
  sleep_time=$(($sleep_time + 5))
done
set -o pipefail
echo "We did it! DNS propagated! 🥳"
