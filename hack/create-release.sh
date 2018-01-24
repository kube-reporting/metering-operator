#!/bin/bash -e

OUTFILE=$1
OUTPUT_DIR_NAME="${OUTFILE%%.zip}"
TMPDIR="$(mktemp -d)"
OUTPUT_DIR="$TMPDIR/$OUTPUT_DIR_NAME"

mkdir -p "$OUTPUT_DIR"

mkdir -p "$OUTPUT_DIR/hack"
cp \
    hack/alm-install.sh \
    hack/alm-uninstall.sh \
    hack/util.sh \
    hack/default-env.sh \
    "$OUTPUT_DIR/hack/"

mkdir -p "$OUTPUT_DIR/Documentation"
cp \
    Documentation/install-chargeback.md \
    Documentation/report.md \
    Documentation/using-chargeback.md \
    Documentation/chargeback-con***REMOVED***g.md \
    Documentation/troubleshooting-chargeback.md \
    Documentation/index.md \
    "$OUTPUT_DIR/Documentation/"

mkdir -p "$OUTPUT_DIR/manifests"
cp -r \
    manifests/custom-resources \
    "$OUTPUT_DIR/manifests/"
cp -r \
    manifests/custom-resource-de***REMOVED***nitions \
    "$OUTPUT_DIR/manifests/"

# Remove scheduled reports folder since we currently do not support them
rm -r "$OUTPUT_DIR/manifests/custom-resources/scheduled-reports"

cp -r \
    manifests/installer \
    "$OUTPUT_DIR/manifests/"

cp -r \
    manifests/alm \
    "$OUTPUT_DIR/manifests/"

cp -r \
    manifests/chargeback-con***REMOVED***g \
    "$OUTPUT_DIR/manifests/"
# Remove minikube values, we don't want users to use this.
rm "$OUTPUT_DIR/tectonic-chargeback-minikube-values.yaml"

echo "Start with Documentation/install-chargeback.md" > "$OUTPUT_DIR/README"

pushd "$TMPDIR"
zip -r "$OUTFILE" "$OUTPUT_DIR_NAME"
popd

mv "$OUTPUT_DIR.zip" .

