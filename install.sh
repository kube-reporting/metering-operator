#!/bin/bash

# Create chargeback namespace
kubectl apply -f manifests/chargeback/namespace.yaml

# Install collection layer
kubectl apply -f manifests/kube-state-metrics # unof***REMOVED***cial build of kube-state-metrics with Node info
kubectl apply -f manifests/promsum
kubectl apply -f manifests/prom-operator

# Install query layer
kubectl apply -f manifests/hive
kubectl apply -f manifests/presto
kubectl apply -f manifests/chargeback
