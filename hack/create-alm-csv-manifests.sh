#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
CHART="$DIR/../charts/chargeback-alm"

# We use cd + pwd in a subshell to turn this into an absolute path (readlink -f isn't cross platform)
OUTPUT_DIR="$(cd "${OUTPUT_DIR:=$DIR/..}" && pwd)"
: "${MANIFESTS_DIR:=$OUTPUT_DIR/manifests}"
: "${INSTALLER_MANIFEST_DIR:=$MANIFESTS_DIR/installer}"
: "${ALM_MANIFEST_DIR:=$MANIFESTS_DIR/alm}"

: "${ALM_VALUES_FILE:=$DIR/chargeback-alm-values.yaml}"

: "${DEPLOYMENT_MANIFEST:=$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-deployment.yaml}"
: "${RBAC_ROLE_MANIFEST:=$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-role.yaml}"
: "${RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST:=$INSTALLER_MANIFEST_DIR/chargeback-helm-operator-service-account.yaml}"
: "${CSV_MANIFEST_DESTINATION:=$ALM_MANIFEST_DIR/chargeback.clusterserviceversion.yaml}"

echo "alm values ***REMOVED***le: $ALM_VALUES_FILE"
echo "Output directory: ${OUTPUT_DIR}"

mkdir -p "${INSTALLER_MANIFEST_DIR}" "${ALM_MANIFEST_DIR}"

JQ_CRD_SCRIPT=$(cat <<EOF
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
***REMOVED***nd "$MANIFESTS_DIR/custom-resource-de***REMOVED***nitions" \
    -type f \
    -exec sh -c 'yamltojson < $0' {} \; \
    | jq -rcs "$JQ_CRD_SCRIPT" > /tmp/alm-crd.json

# Extract the spec of the deployment, and it's name
yamltojson < "$DEPLOYMENT_MANIFEST" \
    | jq -r "$JQ_DEPLOYMENT_SCRIPT" > /tmp/alm-deployment.json


# Slurp both ***REMOVED***les, .[0] is the service account, .[1] is the role, based on the
# arguments ordering. Extracts the rules section of a role, and takes the serviceAccountName
jq \
    -s \
    -r "$JQ_RBAC_SCRIPT" \
    <(yamltojson < "$RBAC_ROLE_SERVICE_ACCOUNT_MANIFEST") \
    <(yamltojson < "$RBAC_ROLE_MANIFEST") > /tmp/alm-permissions.json

# Merge the 3 JSON objects created above
jq -s '.[0] * .[1] * .[2]' \
    /tmp/alm-crd.json \
    /tmp/alm-deployment.json \
    /tmp/alm-permissions.json \
    | jsontoyaml > /tmp/alm-values.yaml

# use helm template to create the csv using our chargeback-alm chart.  the
# chargeback-alm-values is the set of values which are entirely ALM speci***REMOVED***c,
# and the rest are things we can create from our installer manifests and CRD
# manifests.
#
# The sed expression at the end trims trailing whitespace and removes empty
# lines
helm template "$CHART" \
    -f "$ALM_VALUES_FILE" \
    -f /tmp/alm-values.yaml \
    | sed 's/ *$//; /^$/d; /^\s*$/d' \
    > "$CSV_MANIFEST_DESTINATION"
