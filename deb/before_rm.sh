#!/bin/bash

service 'fullerite' stop
service 'fullerite_diamond_server' stop

if [ "$(id 'fullerite')" ]; then
  userdel 'fullerite'
fi

rm -rf /var/log/fullerite
