#!/bin/bash -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$DIR/.."

OUTFILE=$1
BASE_DIR="${2:-$REPO_DIR}"
OUTPUT_DIR_NAME="${OUTFILE%%.zip}"
TMPDIR="$(mktemp -d)"
OUTPUT_DIR="$TMPDIR/$OUTPUT_DIR_NAME"

mkdir -p "$OUTPUT_DIR"
trap "rm -rf $TMPDIR" EXIT

mkdir -p "$OUTPUT_DIR/hack"
cp \
    $BASE_DIR/hack/alm-install.sh \
    $BASE_DIR/hack/alm-uninstall.sh \
    $BASE_DIR/hack/util.sh \
    $BASE_DIR/hack/default-env.sh \
    "$OUTPUT_DIR/hack/"

mkdir -p "$OUTPUT_DIR/Documentation"
cp \
    $BASE_DIR/Documentation/install-chargeback.md \
    $BASE_DIR/Documentation/report.md \
    $BASE_DIR/Documentation/using-chargeback.md \
    $BASE_DIR/Documentation/chargeback-con***REMOVED***g.md \
    $BASE_DIR/Documentation/troubleshooting-chargeback.md \
    $BASE_DIR/Documentation/index.md \
    "$OUTPUT_DIR/Documentation/"

mkdir -p "$OUTPUT_DIR/manifests"
cp -r \
    $BASE_DIR/manifests/custom-resource-de***REMOVED***nitions \
    "$OUTPUT_DIR/manifests/"

cp -r \
    $BASE_DIR/manifests/installer \
    "$OUTPUT_DIR/manifests/"

cp -r \
    $BASE_DIR/manifests/alm \
    "$OUTPUT_DIR/manifests/"

cp -r \
    $BASE_DIR/manifests/chargeback-con***REMOVED***g \
    "$OUTPUT_DIR/manifests/"
# Remove minikube values, we don't want users to use this.
rm "$OUTPUT_DIR/manifests/chargeback-con***REMOVED***g/tectonic-chargeback-values-minikube.yaml"

echo "Start with Documentation/install-chargeback.md" > "$OUTPUT_DIR/README"

pushd "$TMPDIR"
zip -r "$OUTFILE" "$OUTPUT_DIR_NAME"
popd

mv "$OUTPUT_DIR.zip" .

