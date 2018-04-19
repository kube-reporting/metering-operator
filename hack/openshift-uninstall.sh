#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export SKIP_COPY_PULL_SECRET=${SKIP_COPY_PULL_SECRET:=true}
"${DIR}/uninstall.sh"
