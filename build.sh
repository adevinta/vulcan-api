#!/bin/bash

# Copyright 2021 Adevinta

go mod vendor
go generate -v ./...
cd cmd/vulcan-api && GOOS=linux GOARCH=amd64 go build -mod vendor && cd ../..

if [ -n "$IMAGE" ]; then
  docker build . -t "$IMAGE"
  if [ -n "$PUSH_IMAGE" ]; then
    docker push "$IMAGE"
  fi
fi
