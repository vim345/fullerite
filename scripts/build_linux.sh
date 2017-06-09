#!/bin/bash

TAG=$(git describe --abbrev=0 --tags)
if [ ! -z $1 ];then
   TAG=$1
fi

PDIR=$(echo $(pwd) |sed -e 's/scripts//')
docker run -ti -v ${PDIR}:/data/ -w /data/ qnib/golang make
mv ${PDIR}/bin/fullerite ${PDIR}/bin/fullerite-${TAG}-Linux
rm -f bin/glide bin/gocyclo bin/golint bin/beatit

docker run -ti -v ${PDIR}:/data/ -w /data/ qnib/alpn-go-dev make
mv ${PDIR}/bin/fullerite ${PDIR}/bin/fullerite-${TAG}-LinuxMusl
rm -f bin/glide bin/gocyclo bin/golint bin/beatit

