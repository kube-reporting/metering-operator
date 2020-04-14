#! /bin/bash

# Variables:
# IMAGE_OVERRIDE :- when non-empty, overrides the default metering-operator deployment container images
# MANIFEST_DIR :- points to the path to the metering openshift OLM bundle
# MANIFEST_VERSION :- represents the current OLM bundle versioning (e.g. 4.4, 4.5, etc.)

indent() {
  INDENT="      "
  sed "s/^/$INDENT/" | sed "s/^${INDENT}\($1\)/${INDENT:0:-2}- \1/"
}

NAME=${NAME:-"metering-ocp"}
NAMESPACE=${NAMESPACE:-"openshift-metering"}

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
MANIFESTS_DIR="${MANIFESTS_DIR:-${ROOT_DIR}/manifests/deploy/openshift/olm/bundle/}"
MANIFEST_VERSION=${MANIFEST_VERSION:-$(basename $(find $MANIFESTS_DIR -type d | sort -r | head -n 1))}

CRD=$(sed '/^#!.*$/d' ${MANIFESTS_DIR}/${MANIFEST_VERSION}/*crd.yaml | grep -v -- "---" | indent apiVersion)
CSV=$(sed '/^#!.*$/d' ${MANIFESTS_DIR}/${MANIFEST_VERSION}/*version.yaml | sed -e "s,imagePullPolicy: IfNotPresent,imagePullPolicy: Always,"  | sed 's/namespace: placeholder/namespace: '${NAMESPACE}'/' | grep -v -- "---" |  indent apiVersion)
PKG=$(sed '/^#!.*$/d' ${MANIFESTS_DIR}/*package.yaml | indent packageName)

if [ -n "${IMAGE_OVERRIDE:-}" ] ; then
    echo "Overriding the metering-operator deployment images with ${IMAGE_OVERRIDE}"
    CSV=$(echo "${CSV}" | sed -e "s~containerImage:.*~containerImage: ${IMAGE_OVERRIDE}~" | indent apiVersion)
    CSV=$(echo "${CSV}" | sed -e "s~image:.*~image: ${IMAGE_OVERRIDE}\n~" | indent ApiVersion)
fi

cat <<EOF | sed 's/^  *$//' | oc apply -n ${NAMESPACE} -f -
kind: ConfigMap
apiVersion: v1
metadata:
  name: ${NAME}
data:
  customResourceDefinitions: |-
$CRD
  clusterServiceVersions: |-
$CSV
  packages: |-
$PKG
EOF
