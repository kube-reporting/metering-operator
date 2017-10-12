#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
APIS_DIR=$(realpath "${DIR}/../pkg/apis")

***REMOVED***nd "$APIS_DIR" -type f -name '*.go' -not -name 'zz_generated*.go'
