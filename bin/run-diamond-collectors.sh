#!/usr/bin/env bash

BINDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -L)"
FULLERITE_DIR="$(dirname "${BINDIR}")"
SRC_DIR="${FULLERITE_DIR}/src"
EXAMPLE_CONFIG="${FULLERITE_DIR}/fullerite.conf.example"

DIAMOND_DIR="${SRC_DIR}/diamond"
DIAMOND_SERVER="${DIAMOND_DIR}/server.py"

PYTHON="$(which python)"
PYTHONPATH="${SRC_DIR}"
export PYTHONPATH


ARGS="$@"
if [ -z "${ARGS}" ]; then
    ARGS="${EXAMPLE_CONFIG}"
fi

exec $PYTHON ${DIAMOND_SERVER} ${ARGS}
