# Setup local persistence components

## Starting a local database and Kafka

```sh
./persistence-start.sh
./flyway-migrate.sh
```

## Stoping Kafka and Postgres

```sh
./persistence-stop.sh
```

## Cleaning a running Postgres database

```sh
./flyway-clean-schema.sh
```

## Inspecting the local Postgres database

A pgadmin container has been added to easily check the local database content.
It can be accessed just browsing http://locahost:8000 after the local database
has been started. To check the login credentials just see the
`docker-compose.yml` file.
