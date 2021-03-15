#!/usr/bin/env bash
docker run --net="host" -v $(pwd):/scripts boxfuse/flyway -user=vulcan -password=vulcan  -url=jdbc:postgresql://localhost:5432/vulcan -baselineOnMigrate=true -locations=filesystem:/scripts/sql migrate
