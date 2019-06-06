#!/bin/bash
set -e

_date() {
    if [[ "$OSTYPE" == "linux-gnu" ]]; then
        date "$@"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        gdate "$@"
    fi
}

_date "$@"
