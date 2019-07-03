#!/bin/bash

DIR="$(dirname "$0")"

_readlink() {
    if [[ "$OSTYPE" == "linux-gnu" ]]; then
        readlink "$@"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        greadlink "$@"
    ***REMOVED***
}

SUB_MGR_FILE="$(_readlink -f "$DIR/subscription-manager.conf")"
REPO_DIR="$(_readlink -f "$DIR/repos")"

export SUB_MGR_FILE REPO_DIR
