#!/bin/bash
# If $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES and
# $CUSTOM_HELM_OPERATOR_OVERRIDE_VALUES are set, will be used as the paths to
# files containing override values for rendering the manifests to the output
# directory.

set -e
set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

VALUES_ARGS=()

if [[ $# -ge 2 ]] ; then
    METERING_OPERATOR_OUTPUT_DIR=$1
    echo "metering-operator manifest output directory: ${METERING_OPERATOR_OUTPUT_DIR}"
    mkdir -p "${METERING_OPERATOR_OUTPUT_DIR}"
    shift

    OLM_OUTPUT_DIR=$1
    echo "OLM manifest output directory: ${OLM_OUTPUT_DIR}"
    mkdir -p "${OLM_OUTPUT_DIR}"
    shift

    echo "Values files: [$*]"
    # prepends -f to each argument passed in, and stores the list of arguments
    # (-f $arg1 -f $arg2) in VALUES_ARGS
    while (($# > 0)); do
        VALUES_ARGS+=(-f "$1")
        shift
    done
else
    echo "Must specify output directory and values files"
    exit 1
fi

TMPDIR="$(mktemp -d)"
trap 'rm -rf $TMPDIR' EXIT SIGINT
CHART="$ROOT_DIR/charts/metering-ansible-operator"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/operator/role.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-role.yaml"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/operator/rolebinding.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-rolebinding.yaml"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/operator/clusterrole.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-clusterrole.yaml"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/operator/clusterrolebinding.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-clusterrolebinding.yaml"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/operator/deployment.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-deployment.yaml"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/operator/service-account.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/metering-operator-service-account.yaml"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/operator/meteringconfig.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$METERING_OPERATOR_OUTPUT_DIR/meteringconfig.yaml"

# Render the CSV to a temporary location, so we can add the version into it's
# filename after it's been rendered
TMP_CSV="$TMPDIR/metering.clusterserviceversion.yaml"
${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/olm/clusterserviceversion.yaml" \
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

PACKAGE_MANIFEST_DESTINATION="$BUNDLE_DIR/metering.package.yaml"
ART_CONFIG_DESTINATION="$BUNDLE_DIR/art.yaml"

CSV_MANIFEST_DESTINATION="$CSV_BUNDLE_DIR/meteringoperator.v${CSV_VERSION}.clusterserviceversion.yaml"
IMAGE_REFERENCES_MANIFEST_DESTINATION="$CSV_BUNDLE_DIR/image-references"

SUBSCRIPTION_MANIFEST_DESTINATION="$OLM_OUTPUT_DIR/metering.subscription.yaml"
CATALOGSOURCE_MANIFEST_DESTINATION="$OLM_OUTPUT_DIR/metering.catalogsource.yaml"
OPERATORGROUP_MANIFEST_DESTINATION="$OLM_OUTPUT_DIR/metering.operatorgroup.yaml"

mkdir -p "$CSV_BUNDLE_DIR"

# Rename the file with it's version in it, and move it to the final destination
mv -f "$TMP_CSV" "$CSV_MANIFEST_DESTINATION"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/olm/package.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$PACKAGE_MANIFEST_DESTINATION"

# We don't always want to generate an ART package, so check if the ${HELM_BIN} template
# output is empty before redirecting output to a file
HELM_ART_PKG_OUTPUT="$(${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/olm/art.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed")"

if [[ -z "$HELM_ART_PKG_OUTPUT" ]]; then
    echo "Skipping generating an ART package for $BUNDLE_DIR"
else
    echo "$HELM_ART_PKG_OUTPUT" > "$ART_CONFIG_DESTINATION"
fi

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/olm/image-references" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$IMAGE_REFERENCES_MANIFEST_DESTINATION"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/olm/subscription.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$SUBSCRIPTION_MANIFEST_DESTINATION"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/olm/catalogsource.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$CATALOGSOURCE_MANIFEST_DESTINATION"

${HELM_BIN} template "$CHART" \
    ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
    -s "templates/olm/operatorgroup.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$OPERATORGROUP_MANIFEST_DESTINATION"

for CRD_DIR in "$METERING_OPERATOR_OUTPUT_DIR" "$CSV_BUNDLE_DIR"; do
    ${HELM_BIN} template "$CHART" \
        ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
        -s "templates/crds/meteringconfig.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/meteringconfig.crd.yaml"

    ${HELM_BIN} template "$CHART" \
        ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
        -s "templates/crds/report.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/report.crd.yaml"

    ${HELM_BIN} template "$CHART" \
        ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
        -s "templates/crds/reportdatasource.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/reportdatasource.crd.yaml"

    ${HELM_BIN} template "$CHART" \
        ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
        -s "templates/crds/reportquery.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/reportquery.crd.yaml"

    ${HELM_BIN} template "$CHART" \
        ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
        -s "templates/crds/hive.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/hive.crd.yaml"

    ${HELM_BIN} template "$CHART" \
        ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
        -s "templates/crds/prestotable.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/prestotable.crd.yaml"

    ${HELM_BIN} template "$CHART" \
        ${VALUES_ARGS[@]+"${VALUES_ARGS[@]}"} \
        -s "templates/crds/storagelocation.crd.yaml" \
        | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
        > "$CRD_DIR/storagelocation.crd.yaml"
done
