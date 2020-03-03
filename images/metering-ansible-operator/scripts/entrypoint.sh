#!/bin/bash

# add UID to /etc/passwd if missing
if ! whoami &> /dev/null; then
    if [ -w /etc/passwd ]; then
        echo "Adding user ${USER_NAME:-ansible} with current UID $(id -u) to /etc/passwd"
        # Remove existing entry with user first.
        # cannot use sed -i because we do not have permission to write new
        # files into /etc
        sed  "/${USER_NAME:-ansible}:x/d" /etc/passwd > /tmp/passwd
        # add our user with our current user ID into passwd
        echo "${USER_NAME:-ansible}:x:$(id -u):0:${USER_NAME:-hadoop} user:${HOME}:/sbin/nologin" >> /tmp/passwd
        # overwrite existing contents with new contents (cannot replace the
        # file due to permissions)
        cat /tmp/passwd > /etc/passwd
        rm /tmp/passwd
    fi
fi

OPERATOR_SDK_RUN_CMD=${OPERATOR_SDK_RUN_CMD:-"run ansible"}
USE_EXEC_ENTRYPOINT_CMD=${USE_EXEC_ENTRYPOINT_CMD:-false}

if [[ $USE_EXEC_ENTRYPOINT_CMD = true ]]; then
    OPERATOR_SDK_RUN_CMD="exec-entrypoint ansible"
fi

# we expect tini to be in the $PATH
exec tini -- /usr/local/bin/ansible-operator ${OPERATOR_SDK_RUN_CMD} --watches-file=/opt/ansible/watches.yaml "$@"
