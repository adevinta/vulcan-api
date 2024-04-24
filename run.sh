#!/bin/bash

# Copyright 2021 Adevinta

set -e

# export default values for required vars if not set
export PATH_STYLE=${PATH_STYLE:-false}
export COOKIE_NAME=${COOKIE_NAME:-devcon-token}
export DEFAULT_OWNERS=${DEFAULT_OWNERS:-[]}
export DOGSTATSD_ENABLED=${DOGSTATSD_ENABLED:-false}
export AWSCATALOGUE_RETRIES=${AWSCATALOGUE_RETRIES:-4}
export AWSCATALOGUE_RETRY_INTERVAL=${AWSCATALOGUE_RETRY_INTERVAL:-2}
export KAFKA_USER=${KAFKA_USER:-""}
export KAFKA_PASS=${KAFKA_PASS:-""}
export KAFKA_BROKER=${KAFKA_BROKER:-""}
export KAFKA_TOPICS=${KAFKA_TOPICS:-"{}"}
export DNS_HOSTNAME_VALIDATION=${DNS_HOSTNAME_VALIDATION:-true}

envsubst < config.toml > run.toml

# Append global program configuration to run.toml
i=1 GPC_NAME="GPC_${i}_NAME"
while [ -n "${!GPC_NAME}" ]
do
  GPC_ALLOWED_ASSETTYPES="GPC_${i}_ALLOWED_ASSETTYPES"
  GPC_BLOCKED_ASSETTYPES="GPC_${i}_BLOCKED_ASSETTYPES"
  GPC_ALLOWED_CHECKS="GPC_${i}_ALLOWED_CHECKS"
  GPC_BLOCKED_CHECKS="GPC_${i}_BLOCKED_CHECKS"
  GPC_EXCLUDING_SUFFIXES="GPC_${i}_EXCLUDING_SUFFIXES"
  echo "
    [globalpolicy.${!GPC_NAME}]
    allowed_assettypes = ${!GPC_ALLOWED_ASSETTYPES:-[]}
    blocked_assettypes = ${!GPC_BLOCKED_ASSETTYPES:-[]}
    allowed_checks = ${!GPC_ALLOWED_CHECKS:-[]}
    blocked_checks = ${!GPC_BLOCKED_CHECKS:-[]}
    excluding_suffixes = ${!GPC_EXCLUDING_SUFFIXES:-[]}
" >> run.toml
  i=$((i+1))
  GPC_NAME="GPC_${i}_NAME"
done

if [ -n "$PG_CA_B64" ]; then
  mkdir /root/.postgresql
  echo "$PG_CA_B64" | base64 -d > /root/.postgresql/root.crt # for flyway
  echo "$PG_CA_B64" | base64 -d > /etc/ssl/certs/pg.crt # for vulcan-api
fi

flyway -user="$PG_USER" -password="$PG_PASSWORD" \
  -url="jdbc:postgresql://$PG_HOST:$PG_PORT/$PG_NAME?sslmode=$PG_SSLMODE" \
  -baselineOnMigrate=true -locations=filesystem:/app/sql migrate

exec ./vulcan-api -c run.toml
