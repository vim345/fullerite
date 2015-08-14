#!/bin/bash

USER="fullerite"

id $USER > /dev/null 2>&1
if [ $? != 0 ]; then
  useradd --no-create-home --system --user-group $USER
fi

mkdir -p /var/log/fullerite
