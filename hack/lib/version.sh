#!/bin/bash

function load_version_vars() {
    if [[ -n ${METERING_VERSION_FILE-} ]]; then
        source "${METERING_VERSION_FILE}"
        return
    ***REMOVED***
    source "$ROOT_DIR/VERSION"
}
