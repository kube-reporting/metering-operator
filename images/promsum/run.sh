#!/usr/bin/env bash

POLL_INTERVAL=${POLL_INTERVAL:-300}
PROM=${PROM:-http://prometheus.tectonic-system.svc.cluster.local:9090}
SUBJECT=${SUBJECT:-kube-chargeback}

echo "Logging usage data for cluster..."
while true; do
  promsum -prom ${PROM} -before ${POLL_INTERVAL}s

  status=${?}
  if [[ "${status}" != "0" ]]; then
    exit ${status}
  ***REMOVED***

  echo "Waiting ${POLL_INTERVAL} seconds before polling again."
  sleep ${POLL_INTERVAL}
done

