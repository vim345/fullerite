#!/bin/bash

USER="fullerite"

id $USER > /dev/null 2>&1
if [ $? = 0 ]; then
  userdel $USER
fi

rm -rf /var/log/fullerite
