#!/bin/bash -e

set -o pipefail

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

TMPDIR="$(mktemp -d)"

CHART="$ROOT_DIR/charts/metering-alm"

if [[ $# -ge 3 ]] ; then
    DEPLOYER_MANIFESTS_DIR=$1
    echo "Deployer manifest directory: $DEPLOYER_MANIFESTS_DIR"

    DEPLOYMENT_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-operator-deployment.yaml"
    RBAC_ROLE_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-operator-role.yaml"
    RBAC_CLUSTERROLE_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-operator-clusterrole.yaml"
    RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-operator-service-account.yaml"

    for f in "$DEPLOYMENT_MANIFEST" "$RBAC_ROLE_MANIFEST" "$RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST"; do
        if [ ! -e "$f" ]; then
            echo "Expected file $f to exist"
            exit 1
        fi
    done

    shift

    OUTPUT_DIR=$1
    echo "Output directory: ${OUTPUT_DIR}"
    mkdir -p "${OUTPUT_DIR}"
    shift

    echo "ALM values files: $*"
    # prepends -f to each argument passed in, and stores the list of arguments
    # (-f $arg1 -f $arg2) in VALUES_ARGS
    while (($# > 0)); do
        VALUES_ARGS+=(-f "$1")
        shift
    done
else
    echo "Must specify: helm-operator directory, output directory and values files"
    exit 1
fi

mkdir -p "$OUTPUT_DIR"

JQ_CRD_SCRIPT=$(cat <<EOF
sort_by(.metadata.name) |
{
    spec: {
        customresourcedefinitions: {
            owned: map(
                {
                    name: .metadata.name,
                    version: .spec.version,
                    kind: .spec.names.kind,
                    displayName: .metadata.annotations["catalog.app.coreos.com/displayName"],
                    description: .metadata.annotations["catalog.app.coreos.com/description"]
                }
            )
        }
    }
}
EOF
)

JQ_DEPLOYMENT_SCRIPT=$(cat <<EOF
{
    spec: {
        deployments: [
            {
                name: .metadata.name,
                spec: .spec
            }
        ]
    }
}
EOF
)

JQ_RBAC_SCRIPT=$(cat <<EOF
{
    spec: {
        clusterPermissions: [
            {
                serviceAccountName: .[0].metadata.name,
                rules: .[2].rules
            }
        ],
        permissions: [
            {
                serviceAccountName: .[0].metadata.name,
                rules: .[1].rules
            }
        ]
    }
}
EOF
)

# find gets all the CRD files, and execs yamltojson on them, outputting a stream
# of JSON objects separated by newlines.
# this stream of JSON objects is used with jq -s to slurp them into an array,
# which we map over in our $JQ_CRD_SCRIPT.
find "$CRD_DIR" \
    -type f \
    -exec sh -c "$ROOT_DIR/hack/yamltojson < \$0" {} \; \
    | jq -rcs "$JQ_CRD_SCRIPT" > "$TMPDIR/alm-crd.json"

# Extract the spec of the deployment, and it's name
"$ROOT_DIR/hack/yamltojson" < "$DEPLOYMENT_MANIFEST" \
    | jq -r "$JQ_DEPLOYMENT_SCRIPT" > "$TMPDIR/alm-deployment.json"


# Slurp files, which ones are which is based on argument ordering
# .[0] is the ServiceAccount
# .[1] is the Role
# .[2] is the clusterRole
#  Extracts the rules section of a role, and clusterRole, and takes the
#  serviceAccountName
jq \
    -s \
    -r "$JQ_RBAC_SCRIPT" \
    <("$ROOT_DIR/hack/yamltojson" < "$RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST") \
    <("$ROOT_DIR/hack/yamltojson" < "$RBAC_ROLE_MANIFEST") \
    <("$ROOT_DIR/hack/yamltojson" < "$RBAC_CLUSTERROLE_MANIFEST") \
    > "$TMPDIR/alm-permissions.json"

# Merge the 3 JSON objects created above
jq -s '.[0] * .[1] * .[2]' \
    "$TMPDIR/alm-crd.json" \
    "$TMPDIR/alm-deployment.json" \
    "$TMPDIR/alm-permissions.json" \
    | "$ROOT_DIR/hack/jsontoyaml" > "$TMPDIR/alm-values.yaml"

# use helm template to create the csv and package using our metering-alm chart.
# the metering-alm-values is the set of values which are entirely ALM
# specific, and the rest are things we can create from our installer manifests
# and CRD manifests.
#
# The sed expression at the end trims trailing whitespace and removes empty
# lines

# Render the CSV to a temporary location, so we can add the version into it's
# filename after it's been rendered
TMP_CSV="$TMPDIR/metering.clusterserviceversion.yaml"
helm template "$CHART" \
    -f "$TMPDIR/alm-values.yaml" \
    "${VALUES_ARGS[@]}" \
    -x "templates/clusterserviceversion.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$TMP_CSV"

# Rename the file with it's version in it, and move it to the final destination
CSV_VERSION="$("$ROOT_DIR/hack/yamltojson" < "$TMP_CSV" | jq -r '.spec.version' )"
CSV_MANIFEST_DESTINATION="$OUTPUT_DIR/metering.${CSV_VERSION}.clusterserviceversion.yaml"
mv -f "$TMP_CSV" "$CSV_MANIFEST_DESTINATION"

PACKAGE_MANIFEST_DESTINATION="$OUTPUT_DIR/metering.package.yaml"
helm template "$CHART" \
    -f "$TMPDIR/alm-values.yaml" \
    "${VALUES_ARGS[@]}" \
    -x "templates/package.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$PACKAGE_MANIFEST_DESTINATION"

SUBSCRIPTION_MANIFEST_DESTINATION="$OUTPUT_DIR/metering.subscription.yaml"
helm template "$CHART" \
    -f "$TMPDIR/alm-values.yaml" \
    "${VALUES_ARGS[@]}" \
    -x "templates/subscription.yaml" \
    | sed -f "$ROOT_DIR/hack/remove-helm-template-header.sed" \
    > "$SUBSCRIPTION_MANIFEST_DESTINATION"
