#!/bin/bash

function load_version_vars() {
    if [[ -n ${METERING_VERSION_FILE-} ]]; then
        source "${METERING_VERSION_FILE}"
        return
    fi
    source "$ROOT_DIR/VERSION"
}
