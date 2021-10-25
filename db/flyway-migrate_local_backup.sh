#!/usr/bin/env bash

docker run --net=host --rm -v "$PWD":/flyway/sql flyway/flyway:"${FLYWAY_VERSION:-8}-alpine" \
    -user=vulcan -password=vulcan -url=jdbc:postgresql://localhost:5432/vulcan -baselineOnMigrate=true migrate
