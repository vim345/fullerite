#!/usr/bin/env bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -L)"
SRC_DIR="${DIR}/src"
DIAMOND_DIR="${SRC_DIR}/diamond"
DIAMOND_SERVER="${DIAMOND_DIR}/server.py"
PYTHON="$(which python)"
PYTHONPATH="${SRC_DIR}"
export PYTHONPATH

CONFIG_FILE="$1"
if [ -z "${CONFIG_FILE}" ]; then
    CONFIG_FILE="${DIR}/fullerite.conf"
fi

exec $PYTHON ${DIAMOND_SERVER} ${CONFIG_FILE}
