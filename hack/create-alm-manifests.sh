#!/bin/bash -e

set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHART="$DIR/../charts/metering-alm"
: "${MANIFESTS_DIR:=$DIR/../manifests}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-definitions}"

if [[ $# -ge 3 ]] ; then
    DEPLOYER_MANIFESTS_DIR=$1
    echo "Deployer manifest directory: $DEPLOYER_MANIFESTS_DIR"

    DEPLOYMENT_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-helm-operator-deployment.yaml"
    RBAC_ROLE_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-helm-operator-role.yaml"
    RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-helm-operator-service-account.yaml"

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
    -exec sh -c "$DIR/yamltojson < \$0" {} \; \
    | jq -rcs "$JQ_CRD_SCRIPT" > /tmp/alm-crd.json

# Extract the spec of the deployment, and it's name$DIR/
"$DIR/yamltojson" < "$DEPLOYMENT_MANIFEST" \
    | jq -r "$JQ_DEPLOYMENT_SCRIPT" > /tmp/alm-deployment.json


# Slurp both files, .[0] is the service account, .[1] is the role, based on the
# arguments ordering. Extracts the rules section of a role, and takes the serviceAccountName
jq \
    -s \
    -r "$JQ_RBAC_SCRIPT" \
    <("$DIR/yamltojson" < "$RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST") \
    <("$DIR/yamltojson" < "$RBAC_ROLE_MANIFEST") > /tmp/alm-permissions.json

# Merge the 3 JSON objects created above
jq -s '.[0] * .[1] * .[2]' \
    /tmp/alm-crd.json \
    /tmp/alm-deployment.json \
    /tmp/alm-permissions.json \
    | "$DIR/jsontoyaml" > /tmp/alm-values.yaml

# use helm template to create the csv and package using our metering-alm chart.
# the metering-alm-values is the set of values which are entirely ALM
# specific, and the rest are things we can create from our installer manifests
# and CRD manifests.
#
# The sed expression at the end trims trailing whitespace and removes empty
# lines

# Render the CSV to a temporary location, so we can add the version into it's
# filename after it's been rendered
TMP_CSV="/tmp/metering.clusterserviceversion.yaml"
helm template "$CHART" \
    -f /tmp/alm-values.yaml \
    "${VALUES_ARGS[@]}" \
    -x "templates/clusterserviceversion-v1.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$TMP_CSV"

# Rename the file with it's version in it, and move it to the final destination
CSV_VERSION="$("$DIR/yamltojson" < "$TMP_CSV" | jq -r '.spec.version' )"
CSV_MANIFEST_DESTINATION="$OUTPUT_DIR/metering.${CSV_VERSION}.clusterserviceversion.yaml"
mv -f "$TMP_CSV" "$CSV_MANIFEST_DESTINATION"

PACKAGE_MANIFEST_DESTINATION="$OUTPUT_DIR/metering.package.yaml"
helm template "$CHART" \
    -f /tmp/alm-values.yaml \
    "${VALUES_ARGS[@]}" \
    -x "templates/package-v1.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$PACKAGE_MANIFEST_DESTINATION"

SUBSCRIPTION_MANIFEST_DESTINATION="$OUTPUT_DIR/metering.subscription.yaml"
helm template "$CHART" \
    -f /tmp/alm-values.yaml \
    "${VALUES_ARGS[@]}" \
    -x "templates/subscription-v1.yaml" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$SUBSCRIPTION_MANIFEST_DESTINATION"
