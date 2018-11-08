#!/bin/bash

: "${PULL_OSE_METERING_HELM:=true}"

set -x

# rhel
docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/rhel7:7-released'
docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/rhel7:7-released' 'rhel7:7-released'
docker tag 'rhel7:7-released' 'rhel7'
docker tag 'rhel7:7-released' 'rhel'

# golang
docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/golang-builder:1.10'
docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/golang-builder:1.10' openshift/golang-builder:1.10

# openshift base
docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-base:v4.0.0'
docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-base:v4.0.0' openshift/ose-base:v4.0.0

# helm is pulled for building metering-operator
if [ "$PULL_OSE_METERING_HELM" == "true" ]; then
    docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-metering-helm:v4.0'
    docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-metering-helm:v4.0' openshift/ose-metering-helm:v4.0
fi
