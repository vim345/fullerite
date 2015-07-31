#!/usr/bin/env bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -L)"
FULLERITE="${DIR}/bin/fullerite"

ARGS="$@"
if [ -z "${ARGS}" ]; then
    ARGS="-c ${DIR}/fullerite.conf"
fi

exec ${FULLERITE} ${ARGS}
