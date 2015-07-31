#!/usr/bin/env bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -L)"
FULLERITE="${DIR}/bin/fullerite"

CONFIG_FILE="$1"
if [ -z "${CONFIG_FILE}" ]; then
    CONFIG_FILE="${DIR}/fullerite.conf"
fi

exec ${FULLERITE} -c ${CONFIG_FILE}
