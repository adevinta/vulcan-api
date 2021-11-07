#!/bin/bash

# exit on error
set -e

docker stop keycloak vulcanito_db pgadmin
docker rm keycloak vulcanito_db pgadmin
