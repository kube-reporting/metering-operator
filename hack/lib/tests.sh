#!/bin/bash -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/../..
source "${ROOT_DIR}/hack/common.sh"

# IMAGE_FORMAT comes from ci-operator https://github.com/openshift/ci-operator/blob/master/TEMPLATES.md#image_format
# TEST_IMAGE_REPO is the image repo, which is everything before the ":${component}" value in
# registry.svc.ci.openshift.org/ci-op-<input-hash>/stable:${component}
IMAGE_FORMAT="${IMAGE_FORMAT:-}"
TEST_IMAGE_REPO="${IMAGE_FORMAT%:*}"
export METERING_OPERATOR_DEPLOY_REPO="${METERING_OPERATOR_DEPLOY_REPO:-$TEST_IMAGE_REPO}"
export REPORTING_OPERATOR_DEPLOY_REPO="${REPORTING_OPERATOR_DEPLOY_REPO:-$TEST_IMAGE_REPO}"

# image tags are the ${component} in the $IMAGE_FORMAT: registry.svc.ci.openshift.org/ci-op-<input-hash>/stable:${component}
# for metering-operator and reporting-operator being tested in ci, these are unchanging
export METERING_OPERATOR_DEPLOY_TAG="${METERING_OPERATOR_DEPLOY_TAG:-"metering-operator"}"
export REPORTING_OPERATOR_DEPLOY_TAG="${REPORTING_OPERATOR_DEPLOY_TAG:-"reporting-operator"}"
