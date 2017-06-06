#!/bin/bash

hadoop fs -mkdir -p /pod-data

POLL_INTERVAL=${POLL_INTERVAL:-300}

id=$(pod-data -id)

echo "Logging usage data for cluster ${id}"
while true; do
  ***REMOVED***le="$(date +%s).json"
  pod-data s3:///coreos-team-chargeback/k8s-usage/${id}/${***REMOVED***le}
  echo "Waiting ${POLL_INTERVAL} seconds before polling again."
  sleep ${POLL_INTERVAL}
done

