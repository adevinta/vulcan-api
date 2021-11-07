#!/bin/bash

# exit on error
set -e

cd ../../cmd/vulcan-api

go build

cd -

../../cmd/vulcan-api/vulcan-api -c $1
