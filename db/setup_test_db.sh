psql -h localhost -c "CREATE DATABASE vulcanito_test;" -U vulcan vulcan
psql -h localhost -c "CREATE USER vulcanito_test WITH PASSWORD 'vulcanito_test';" -U vulcan vulcan
psql -h localhost -c "ALTER USER vulcanito_test WITH SUPERUSER;" -U vulcan vulcan
db/flyway/flyway -user=vulcanito_test -password=vulcanito_test -url=jdbc:postgresql://localhost:5432/vulcanito_test -baselineOnMigrate=true -locations=filesystem:db/sql,filesystem:db/test-sql migrate
psql -h localhost -c "CREATE DATABASE vulcanito WITH TEMPLATE vulcanito_test OWNER vulcanito_test;" -U vulcan vulcan

