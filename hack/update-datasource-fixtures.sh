#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..

if [ $# -ne 5 ]; then
    echo "usage: reporting_operator_base_url namespace data_start data_end out_dir"
    exit 1
***REMOVED***

base_url=$1
namespace=$2
data_start=$3
data_end=$4
out_dir=$5

DATASOURCES="$(kubectl -n "$namespace" get reportdatasources -o name | cut -d/ -f2)"

echo "getting metrics for ${DATASOURCES[*]}"

while read -r ds; do
    if [ -n "$ds" ]; then
        url="$base_url/api/v1/datasources/prometheus/fetch/$namespace/$ds?start=$data_start&end=$data_end"
        echo "fetching results from $url"
        curl -k "$url" | faq -f json -o json -M -r '.' > "$out_dir/$ds.json"
    ***REMOVED***
done <<< "$DATASOURCES"
