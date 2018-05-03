#!/bin/bash -e

set -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHART="$DIR/../charts/metering-alm"
: "${MANIFESTS_DIR:=$DIR/../manifests}"
: "${CRD_DIR:=$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions}"

if [[ $# -ge 3 ]] ; then
    DEPLOYER_MANIFESTS_DIR=$1
    echo "Deployer manifest directory: $DEPLOYER_MANIFESTS_DIR"

    DEPLOYMENT_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-helm-operator-deployment.yaml"
    RBAC_ROLE_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-helm-operator-role.yaml"
    RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST="$DEPLOYER_MANIFESTS_DIR/metering-helm-operator-service-account.yaml"

    for f in "$DEPLOYMENT_MANIFEST" "$RBAC_ROLE_MANIFEST" "$RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST"; do
        if [ ! -e "$f" ]; then
            echo "Expected ***REMOVED***le $f to exist"
            exit 1
        ***REMOVED***
    done

    shift

    OUTPUT_DIR=$1
    echo "Output directory: ${OUTPUT_DIR}"
    mkdir -p "${OUTPUT_DIR}"
    CSV_MANIFEST_DESTINATION="$OUTPUT_DIR/metering.clusterserviceversion.yaml"
    shift

    echo "ALM values ***REMOVED***les: $*"
    # prepends -f to each argument passed in, and stores the list of arguments
    # (-f $arg1 -f $arg2) in VALUES_ARGS
    while (($# > 0)); do
        VALUES_ARGS+=(-f "$1")
        shift
    done
***REMOVED***
    echo "Must specify: helm-operator directory, output directory and values ***REMOVED***les"
    exit 1
***REMOVED***

mkdir -p "$OUTPUT_DIR"

JQ_CRD_SCRIPT=$(cat <<EOF
sort_by(.metadata.name) |
{
    spec: {
        customresourcede***REMOVED***nitions: {
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

# ***REMOVED***nd gets all the CRD ***REMOVED***les, and execs yamltojson on them, outputting a stream
# of JSON objects separated by newlines.
# this stream of JSON objects is used with jq -s to slurp them into an array,
# which we map over in our $JQ_CRD_SCRIPT.
***REMOVED***nd "$CRD_DIR" \
    -type f \
    -exec sh -c "$DIR/yamltojson < \$0" {} \; \
    | jq -rcs "$JQ_CRD_SCRIPT" > /tmp/alm-crd.json

# Extract the spec of the deployment, and it's name$DIR/
"$DIR/yamltojson" < "$DEPLOYMENT_MANIFEST" \
    | jq -r "$JQ_DEPLOYMENT_SCRIPT" > /tmp/alm-deployment.json


# Slurp both ***REMOVED***les, .[0] is the service account, .[1] is the role, based on the
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

# use helm template to create the csv using our metering-alm chart.
# the metering-alm-values is the set of values which are entirely ALM
# speci***REMOVED***c, and the rest are things we can create from our installer manifests
# and CRD manifests.
#
# The sed expression at the end trims trailing whitespace and removes empty
# lines
helm template "$CHART" \
    "${VALUES_ARGS[@]}" \
    | sed -f "$DIR/remove-helm-template-header.sed" \
    > "$CSV_MANIFEST_DESTINATION"
