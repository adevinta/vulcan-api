docker run \
--health-cmd='curl --fail http://$(hostname -i):8080/auth/realms/vulcan' \
--health-interval=2s \
--name=keycloak \
--rm \
-d \
-p 8093:8080 -v $(pwd)/vulcan-realm:/tmp/vulcan-realm -e KEYCLOAK_USER=admin -e KEYCLOAK_PASSWORD=admin quay.io/keycloak/keycloak:15.0.2 -Dkeycloak.migration.action=import -Dkeycloak.migration.provider=dir -Dkeycloak.migration.dir=/tmp/vulcan-realm
