#!/bin/bash

: "${PULL_OSE_METERING_HELM:=true}"
: "${PULL_OSE_ANSIBLE_OPERATOR:=true}"

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
docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-base:latest'
docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-base:latest' openshift/ose-base:latest

# openshift cli
docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-cli:latest'
docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-cli:latest' openshift/ose-cli:latest

# helm is pulled for building metering-operator
if [ "$PULL_OSE_METERING_HELM" == "true" ]; then
    docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-metering-helm:latest'
    docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-metering-helm:latest' openshift/ose-metering-helm:latest
***REMOVED***

# ansible-operator is pulled for building metering-ansible-operator
if [ "$PULL_OSE_ANSIBLE_OPERATOR" == "true" ]; then
    docker pull 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-ansible-operator:latest'
    docker tag 'brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888/openshift/ose-ansible-operator:latest' openshift/ose-ansible-operator:latest
***REMOVED***
