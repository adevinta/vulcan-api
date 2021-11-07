#!/bin/bash

# exit on error
set -e

echo "Setting up Keycloak...
"

cd ../keycloak/

./start-keycloak.sh

while [ "`docker inspect -f {{.State.Health.Status}} keycloak`" != "healthy" ]; do     
    sleep 2; 
done

echo "--> Keycloak ready.

Setting up postgres...
"

cd ../../db

./postgres-start.sh

while [ "`docker inspect -f {{.State.Health.Status}} vulcanito_db`" != "healthy" ]; do     
    sleep 2; 
done

echo "--> Postgres ready.

Applying migrations...
"

./flyway-migrate.sh

echo "--> Migrations applied successfully.

--> You can start the Vulcan API now.
"
