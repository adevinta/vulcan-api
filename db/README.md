# Starting a local database

```
$ postgres-start.sh
$ flyway-migrate.sh
```

# Stoping the database

```
$ postgres-stop.sh
```

# Cleaning a running database

```
$ flyway-clean-schema.sh
```

# Inspecting the database

A pgadmin container has been added to easily check the local database content. It can be accessed just browsing http://locahost:8000 after the local database has been started. To check the login credentials just see the `docker-compose.yml` file.
