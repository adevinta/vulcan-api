#!/usr/bin/env bash
docker run --net="host" -v $(pwd):/scripts boxfuse/flyway -user=vulcanito      -password=vulcanito      -url=jdbc:postgresql://localhost:5432/vulcanito      -baselineOnMigrate=true -locations=filesystem:/scripts/ clean
docker run --net="host" -v $(pwd):/scripts boxfuse/flyway -user=vulcanito_test -password=vulcanito_test -url=jdbc:postgresql://localhost:5432/vulcanito_test -baselineOnMigrate=true -locations=filesystem:/scripts/ clean
#docker exec vulcanito_db psql -c "DROP USER vulcanito_test;" -U postgres
