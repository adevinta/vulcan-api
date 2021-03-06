# vulcan-api

## Requirements

```
go generate ./...
```

## Running

```
cd cmd/vulcan-api
go install
vulcan-api -c ../../_resources/config/local.toml
```

# Docker execute

Those are the variables you have to setup:

|Variable|Description|Sample|
|---|---|---|
|PORT||8080|
|SECRET_KEY||SUPERSECRETSIGNKEY|
|COOKIE_DOMAIN||localhost|
|PG_HOST||localhost|
|PG_NAME||vulcanito|
|PG_USER||vulcanito|
|PG_PASSWORD||vulcanito|
|PG_PORT||5432|
|PG_SSLMODE|One of these (disable,allow,prefer,require,verify-ca,verify-full)|disable|
|PG_CA_B64|A base64 encoded ca certificate||
|LOG_LEVEL| ERROR, WARN, INFO or DEBUG (default INFO)|
|SAML_MEATADATA|||
|SAML_ISSUER|||
|SAML_CALLBACK||http://localhost:8080/api/v1/login/callback|
|SAML_TRUSTED_DOMAINS||["localhost"]|
|SCANENGINE_URL||http://localhost:8081/v1/|
|SCHEDULER_URL||http://localhost:8082/|
|SQS_POLLING_INTERVAL||10|
|SQS_WAIT_TIME||20|
|SQS_TIMEOUT||3600|
|SQS_QUEUE_ARN||arn:aws:sqs:xxx:123456789012:yyy|
|AWS_SQS_ENDPOINT|Optional||
|REPORTS_SNS_ARN||arn:aws:sns:xxx:123456789012:yyy|
|AWS_SNS_ENDPOINT|Optional||
|REPORTS_API_URL||http://localhost:8084|
|PERSISTENCE_HOST||persistence.vulcan.example.com|
|VULNERABILITYDB_URL||http://localhost:8083|
|SCAN_REDIRECT_URL|Redirecting URL for reports, OPTIONAL|https://insights-redirect.vulcan.s3-xxx.amazonaws.com/index.html?reportUrl=|
|VULCAN_UI_URL|Vulcan UI base URL for Digest report link|http://localhost:1234|

First we have to build the `vulcan-api` because the build only copies the file.

We need to provide `linux` compiled binary to the docker build command. This won't be necessary when this component has been open sourced.
For now, we need to do some extra steps:

```bash
./build.sh

docker build . -t va

# Use the default config.toml customized with env variables.
docker run --env-file ./local.env va

# Or set the env variables one by one....
docker run --env PORT=8888  .........    ./local.env va

# Use custom config.toml
docker run -v `pwd`/custom.toml:/app/config.toml va
```
