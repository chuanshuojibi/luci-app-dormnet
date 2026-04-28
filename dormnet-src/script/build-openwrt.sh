#!/bin/bash

cd $(dirname $0)/../

PACKAGE=github.com/openwrt-dormnet/dormnet
CURRENT_TIME=$(date -u -Iseconds)

desc=$(git describe --tags --long --always)
VERSION_NAME=${desc%%-*}
count=$(echo "${desc}" | awk -F- '{print $(NF-1)}')
if [ "${count}" != "0" ]; then
  VERSION_NAME="${VERSION_NAME}+${count}"
fi

export CGO_ENABLED=0

go generate ./...
go build \
  -trimpath \
  -buildvcs=false \
  -ldflags="-s -w -X '${PACKAGE}/internal.constants.version=${VERSION_NAME}‘ -X ‘${PACKAGE}/internal.constants.buildTime=${CURRENT_TIME}'" \
  -o dormnet ./cmd/dormnet
