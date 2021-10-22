#!/bin/sh

# Copyright 2021 Adevinta

# export default values for required vars if not set
export PATH_STYLE=${PATH_STYLE:-false}
export COOKIE_NAME=${COOKIE_NAME:-devcon-token}
export DEFAULT_OWNERS=${DEFAULT_OWNERS:-[]}
export SQS_POLLING_INTERVAL=${SQS_POLLING_INTERVAL:-10}
export SQS_WAIT_TIME=${SQS_WAIT_TIME:-20}
export SQS_TIMEOUT=${SQS_TIMEOUT:-3600}
export DOGSTATSD_ENABLED=${DOGSTATSD_ENABLED:-false}
export AWSCATALOGUE_RETRIES=${AWSCATALOGUE_RETRIES:-4}
export AWSCATALOGUE_RETRY_INTERVAL=${AWSCATALOGUE_RETRY_INTERVAL:-2}

cat config.toml | envsubst > run.toml

if [ ! -z "$PG_CA_B64" ]; then
  mkdir /root/.postgresql
  echo $PG_CA_B64 | base64 -d > /root/.postgresql/root.crt   # for flyway
  echo $PG_CA_B64 | base64 -d > /etc/ssl/certs/pg.crt  # for vulcan-api
fi

flyway -user=$PG_USER -password=$PG_PASSWORD \
  -url=jdbc:postgresql://$PG_HOST:$PG_PORT/$PG_NAME?sslmode=$PG_SSLMODE \
  -community -baselineOnMigrate=true -locations=filesystem:/app/sql migrate

./vulcan-api -c run.toml
