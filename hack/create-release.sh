#!/bin/bash -e

OUTFILE=$1
OUTDIR_NAME="${OUTFILE%%.zip}"
TEMPDIR="$(mktemp -d)/$OUTDIR_NAME"

mkdir -p "$TEMPDIR"

mkdir -p "$TEMPDIR/hack"
cp \
    hack/install.sh \
    hack/alm-install.sh \
    hack/install.sh \
    hack/uninstall.sh \
    hack/util.sh \
    "$TEMPDIR/hack/"

mkdir -p "$TEMPDIR/Documentation"
cp \
    Documentation/Installation.md \
    Documentation/Report.md \
    Documentation/Using-chargeback.md \
    "$TEMPDIR/Documentation/"

mkdir -p "$TEMPDIR/manifests"
cp -r \
    manifests/custom-resources \
    "$TEMPDIR/manifests/"

# Remove scheduled reports folder since we currently do not support them
rm -r "$TEMPDIR/manifests/custom-resources/scheduled-reports"

cp -r \
    manifests/installer \
    "$TEMPDIR/manifests/"

cp -r \
    manifests/alm \
    "$TEMPDIR/manifests/"

mkdir -p $TEMPDIR/manifests/chargeback-config
cp \
    manifests/chargeback-config/custom-values.yaml \
    "$TEMPDIR/manifests/chargeback-config"
echo "Start with Documentation/Installation.md" > "$TEMPDIR/README"

zip -r "$OUTFILE" "$TEMPDIR"
