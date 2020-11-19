#! /bin/bash

set -eou pipefail
set -x

CATALOG_SOURCE_NAMESPACE=${CATALOG_SOURCE_NAMESPACE:=openshift-marketplace}
CATALOG_SOURCE_NAME=${CATALOG_SOURCE_NAME:=redhat-operators}

OUTPUT_DIRECTORY=${OUTPUT_DIRECTORY:=$(mktemp -ud)/resources}
mkdir -p "${OUTPUT_DIRECTORY}"

oc -n ${CATALOG_SOURCE_NAMESPACE} get catalogsources ${CATALOG_SOURCE_NAME} -o yaml > "${OUTPUT_DIRECTORY}"/catalogsource.yaml
oc -n ${CATALOG_SOURCE_NAMESPACE} get packagemanifests > "${OUTPUT_DIRECTORY}"/packagemanifests.yaml
oc -n ${CATALOG_SOURCE_NAMESPACE} get packagemanifests -l "catalog-namespace=${CATALOG_SOURCE_NAMESPACE},catalog=${CATALOG_SOURCE_NAME}" -o yaml > "${OUTPUT_DIRECTORY}"/metering-ocp-packagemanifests.yaml
