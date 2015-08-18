#!/bin/bash

if [ "$(id 'fullerite')" ]; then
  echo "creating user fullerite"
  useradd --no-create-home --system --user-group "fullerite"
fi

echo "creating log directory: /var/log/fullerite"
mkdir -p /var/log/fullerite
chown fullerite:fullerite /var/log/fullerite
