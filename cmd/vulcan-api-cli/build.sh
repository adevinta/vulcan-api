#!/bin/bash

# Copyright 2021 Adevinta

set -e

# Autogenerate content
go run ./gen/main.go

# Compile and install
go install ./tool/vulcan-api-cli

cp -p ../../swagger/swagger.json ../../docs/swagger.json

jq < ../../docs/swagger.json > ../../swagger/swagger.json
