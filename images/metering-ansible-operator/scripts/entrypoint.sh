#!/bin/bash

# we expect tini to be in the $PATH
exec tini -- /usr/local/bin/ansible-operator run ansible --watches-file=/opt/ansible/watches.yaml "$@"
