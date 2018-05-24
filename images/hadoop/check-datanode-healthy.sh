#!/bin/bash

: "${DATANODE_ADDRESS:=127.0.0.1:50075}"

set -ex

if [ "$(curl "$DATANODE_ADDRESS/jmx?qry=Hadoop:service=DataNode,name=DataNodeInfo" | jq '.beans[0].NamenodeAddresses' -r | jq 'to_entries | map(.value) | all')" == "true" ]; then
    echo "Name node addresses all have addresses, healthy"
    exit 0
else
    echo "found null namenode addresses in JMX metrics, unhealthy"
    exit 1
fi
