#!/bin/bash
set -e
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

: "${1?"Usage: $0 IMAGE_TAG"}"

TMPDIR="$(mktemp -d)"

trap "rm -rf $TMPDIR" EXIT

"$DIR/render-chargeback-helm-operator-override-values.sh" "$1" > "$TMPDIR/override-helm-operator-values.yaml"
"$DIR/render-chargeback-alm-override-values.sh" "$1" > "$TMPDIR/override-alm-values.yaml"
"$DIR/create-installer-manifests.sh" "$TMPDIR/override-helm-operator-values.yaml"
"$DIR/create-alm-csv-manifests.sh" "$TMPDIR/override-alm-values.yaml"
