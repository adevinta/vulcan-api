#!/usr/bin/env bash

docker run -q --net=host --rm -v "$PWD":/flyway/sql flyway/flyway:"${FLYWAY_VERSION:-10}-alpine" \
    -user=vulcanito -password=vulcanito -url=jdbc:postgresql://localhost:5432/vulcanito -baselineOnMigrate=true -cleanDisabled=false clean

docker run -q --net=host --rm -v "$PWD":/flyway/sql flyway/flyway:"${FLYWAY_VERSION:-10}-alpine" \
    -user=vulcanito_test -password=vulcanito_test -url=jdbc:postgresql://localhost:5432/vulcanito_test -baselineOnMigrate=true -cleanDisabled=false clean

#docker exec vulcanito_db psql -c "DROP USER vulcanito_test;" -U postgres
