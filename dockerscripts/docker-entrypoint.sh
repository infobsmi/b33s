#!/bin/sh
#

# If command starts with an option, prepend b33s.
if [ "${1}" != "b33s" ]; then
    if [ -n "${1}" ]; then
        set -- b33s "$@"
    fi
fi

# su-exec to requested user, if service cannot run exec will fail.
docker_switch_user() {
    if [ -n "${B33S_USERNAME}" ] && [ -n "${B33S_GROUPNAME}" ]; then
        if [ -n "${B33S_UID}" ] && [ -n "${B33S_GID}" ]; then
            groupadd -g "$B33S_GID" "$b33s_GROUPNAME" && \
                useradd -u "$b33s_UID" -g "$b33s_GROUPNAME" "$b33s_USERNAME"
        else
            groupadd "$b33s_GROUPNAME" && \
                useradd -g "$b33s_GROUPNAME" "$b33s_USERNAME"
        fi
        exec setpriv --reuid="${B33S_USERNAME}" \
             --regid="${B33S_GROUPNAME}" --keep-groups "$@"
    else
        exec "$@"
    fi
}

## Switch to user if applicable.
docker_switch_user "$@"
