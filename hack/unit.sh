#!/bin/bash
set -e

ROOT_DIR=$(dirname "${BASH_SOURCE}")/..
source "${ROOT_DIR}/hack/common.sh"

: "${TEST_OUTPUT_DIR:="."}"
: "${COVERAGE_OUTFILE:="$TEST_OUTPUT_DIR/coverage.out"}"
: "${JUNIT_REPORT_OUTFILE:="$TEST_OUTPUT_DIR/junit-metering.xml"}"

TMP_DIR="$(mktemp -d)"

trap "rm -rf $TMP_DIR" exit

mkdir -p "$TEST_OUTPUT_DIR"
go test -v -coverprofile="$COVERAGE_OUTFILE" ./pkg/... 2>&1 | tee "$TMP_DIR/metering-test-output.txt"
if command -v go-junit-report >/dev/null 2>&1; then
    go-junit-report < "$TMP_DIR/metering-test-output.txt" > "${JUNIT_REPORT_OUTFILE}"
fi
go test -c -o bin/e2e-tests ./test/e2e
go test -c -o bin/integration-tests ./test/integration
