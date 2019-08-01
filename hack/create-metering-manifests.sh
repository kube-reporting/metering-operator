#!/bin/bash
# If $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES and
# $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES are set, will be used as the paths to
# ***REMOVED***les containing override values for rendering the manifests to the output
# directory.

set -e
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"


if [[ $# -ge 3 ]] ; then
    METERING_OPERATOR_OUTPUT_DIR=$1
    echo "metering-operator manifest output directory: ${METERING_OPERATOR_OUTPUT_DIR}"
    mkdir -p "${METERING_OPERATOR_OUTPUT_DIR}"
    shift

    OLM_OUTPUT_DIR=$1
    echo "OLM manifest output directory: ${OLM_OUTPUT_DIR}"
    mkdir -p "${OLM_OUTPUT_DIR}"
    shift

    echo "Values ***REMOVED***les: $*"
    # prepends -f to each argument passed in, and stores the list of arguments
    # (-f $arg1 -f $arg2) in VALUES_ARGS
    while (($# > 0)); do
        VALUES_ARGS+=(-f "$1")
        shift
    done
***REMOVED***
    echo "Must specify output directory and values ***REMOVED***les"
    exit 1
***REMOVED***

TMPDIR="$(mktemp -d)"
trap 'rm -rf $TMPDIR' EXIT SIGINT
CHART="$ROOT_DIR/charts/metering-ansible-operator"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/operator/role.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-role.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/operator/rolebinding.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-rolebinding.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/operator/clusterrole.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-clusterrole.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/operator/clusterrolebinding.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-clusterrolebinding.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/operator/deployment.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-deployment.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/operator/service-account.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-service-account.yaml"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/operator/meteringcon***REMOVED***g.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/meteringcon***REMOVED***g.yaml"

# Render the CSV to a temporary location, so we can add the version into it's
# ***REMOVED***lename after it's been rendered
TMP_CSV="$TMPDIR/metering.clusterserviceversion.yaml"
helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/olm/clusterserviceversion.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$TMP_CSV"

# extract the CSV version
CSV_VERSION="$("$FAQ_BIN" -M -c -r -o json '.spec.version' "$TMP_CSV" )"
# get the major.minor from a semver
# shellcheck disable=SC2206
semver=( ${CSV_VERSION//./ } )
major="${semver[0]}"
minor="${semver[1]}"

# the root directory containing the package.yaml
BUNDLE_DIR="$OLM_OUTPUT_DIR/bundle"
# the versioned directory containing CSVs and CRDs for each major.minor version
CSV_BUNDLE_DIR="$BUNDLE_DIR/${major}.${minor}"

PACKAGE_MANIFEST_DESTINATION="$BUNDLE_DIR/package.yaml"
ART_CONFIG_DESTINATION="$BUNDLE_DIR/art.yaml"

CSV_MANIFEST_DESTINATION="$CSV_BUNDLE_DIR/meteringoperator.v${CSV_VERSION}.clusterserviceversion.yaml"
IMAGE_REFERENCES_MANIFEST_DESTINATION="$CSV_BUNDLE_DIR/image-references"

SUBSCRIPTION_MANIFEST_DESTINATION="$OLM_OUTPUT_DIR/metering.subscription.yaml"
CATALOGSOURCECONFIG_MANIFEST_DESTINATION="$OLM_OUTPUT_DIR/metering.catalogsourcecon***REMOVED***g.yaml"
OPERATORGROUP_MANIFEST_DESTINATION="$OLM_OUTPUT_DIR/metering.operatorgroup.yaml"

mkdir -p "$CSV_BUNDLE_DIR"

# Rename the ***REMOVED***le with it's version in it, and move it to the ***REMOVED***nal destination
mv -f "$TMP_CSV" "$CSV_MANIFEST_DESTINATION"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/olm/package.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$PACKAGE_MANIFEST_DESTINATION"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/olm/art.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$ART_CONFIG_DESTINATION"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/olm/image-references" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$IMAGE_REFERENCES_MANIFEST_DESTINATION"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/olm/subscription.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$SUBSCRIPTION_MANIFEST_DESTINATION"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/olm/catalogsourcecon***REMOVED***g.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$CATALOGSOURCECONFIG_MANIFEST_DESTINATION"

helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    -x "templates/olm/operatorgroup.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$OPERATORGROUP_MANIFEST_DESTINATION"

for CRD_DIR in "$METERING_OPERATOR_OUTPUT_DIR" "$CSV_BUNDLE_DIR"; do
    helm template "$CHART" \
        "${VALUES_ARGS[@]}" \
        -x "templates/crds/meteringcon***REMOVED***g.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/meteringcon***REMOVED***g.crd.yaml"

    helm template "$CHART" \
        "${VALUES_ARGS[@]}" \
        -x "templates/crds/report.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/report.crd.yaml"

    helm template "$CHART" \
        "${VALUES_ARGS[@]}" \
        -x "templates/crds/reportdatasource.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/reportdatasource.crd.yaml"

    helm template "$CHART" \
        "${VALUES_ARGS[@]}" \
        -x "templates/crds/reportquery.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/reportquery.crd.yaml"

    helm template "$CHART" \
        "${VALUES_ARGS[@]}" \
        -x "templates/crds/hive.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/hive.crd.yaml"

    helm template "$CHART" \
        "${VALUES_ARGS[@]}" \
        -x "templates/crds/prestotable.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/prestotable.crd.yaml"

    helm template "$CHART" \
        "${VALUES_ARGS[@]}" \
        -x "templates/crds/storagelocation.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/storagelocation.crd.yaml"
done

