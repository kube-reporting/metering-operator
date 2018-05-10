#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

# This will be overriden by init.sh, but is needed to properly find init.sh and
# source it
ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd -P)

# Sets up all common environment variables/functions
# shellcheck source=hack/lib/init.sh
source "${ROOT_DIR}/hack/lib/init.sh"
