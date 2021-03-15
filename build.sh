#!/bin/bash

# Copyright 2021 Adevinta

go mod vendor
go generate -v ./...
cd cmd/vulcan-api && GOOS=linux GOARCH=amd64 go build -mod vendor && cd ../..

image=$(echo $IMAGE)

if [ ! -z "$image" ]; then
  docker build . -t $IMAGE
  if $PUSH_IMAGE
  then
    docker push $image
  fi
fi
