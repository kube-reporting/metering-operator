#!/bin/bash

set -eou pipefail

echo -e "Dumping IMAGE env vars\n"
env | grep IMAGE
echo -e "\n\n"

# Need to handle the case where the metering and reporting operator images have already been set
# Right now, we're depending on the METERING_OPERATOR_IMAGE_REPO and METERING_OPERATOR_IMAGE_TAG
# (and the same for the reporting-operator) so we need to remain consistent with that approach
IMAGE_METERING_ANSIBLE_OPERATOR=${IMAGE_METERING_ANSIBLE_OPERATOR:-registry.svc.ci.openshift.org/ocp/4.7:metering-ansible-operator}
IMAGE_METERING_REPORTING_OPERATOR=${IMAGE_METERING_REPORTING_OPERATOR:-registry.svc.ci.openshift.org/ocp/4.7:metering-reporting-operator}
IMAGE_METERING_PRESTO=${IMAGE_METERING_PRESTO:-registry.svc.ci.openshift.org/ocp/4.7:metering-presto}
IMAGE_METERING_HIVE=${IMAGE_METERING_HIVE:-registry.svc.ci.openshift.org/ocp/4.7:metering-hive}
IMAGE_METERING_HADOOP=${IMAGE_METERING_HADOOP:-registry.svc.ci.openshift.org/ocp/4.7:metering-hadoop}
IMAGE_GHOSTUNNEL=${IMAGE_GHOSTUNNEL:-registry.svc.ci.openshift.org/ocp/4.7:ghostunnel}
IMAGE_OAUTH_PROXY=${IMAGE_OAUTH_PROXY:-registry.svc.ci.openshift.org/ocp/4.7:oauth-proxy}

# update the manifest with the image built by ci
sed -i "s,quay.io/openshift/origin-metering-ansible-operator:4.7,${IMAGE_METERING_ANSIBLE_OPERATOR}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-metering-reporting-operator:4.7,${IMAGE_METERING_REPORTING_OPERATOR}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-metering-presto:4.7,${IMAGE_METERING_PRESTO}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-metering-hive:4.7,${IMAGE_METERING_HIVE}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-metering-hadoop:4.7,${IMAGE_METERING_HADOOP}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-ghostunnel:4.7,${IMAGE_GHOSTUNNEL}," /manifests/*/*clusterserviceversion.yaml
sed -i "s,quay.io/openshift/origin-oauth-proxy:4.7,${IMAGE_OAUTH_PROXY}," /manifests/*/*clusterserviceversion.yaml

echo -e "substitution complete, dumping new csv\n\n"
cat /manifests/*/*clusterserviceversion.yaml

echo "Generating the sqlite database"
/usr/bin/initializer --manifests=/manifests --output=/bundle/bundles.db --permissive=true
