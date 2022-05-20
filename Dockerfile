# Copyright 2021 Adevinta

FROM golang:1.18.2-alpine3.15 as builder

WORKDIR /app

COPY . .

RUN cd cmd/vulcan-api && GOOS=linux GOARCH=amd64 go build . && cd -

FROM alpine:3.15

WORKDIR /flyway

RUN apk add --no-cache --update openjdk8-jre-base bash gettext libc6-compat

ARG FLYWAY_VERSION=8.5.9

RUN wget https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && tar -xzf flyway-commandline-${FLYWAY_VERSION}.tar.gz --strip 1 \
    && rm flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && find ./drivers/ -type f -not -name 'postgres*' -delete \
    && chown -R root:root . \
    && ln -s /flyway/flyway /bin/flyway

ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"

ENV BUILD_RFC3339 "$BUILD_RFC3339"
ENV COMMIT "$COMMIT"

WORKDIR /app

COPY db/sql /app/sql/

RUN mkdir -p /app/output

COPY --from=builder /app/cmd/vulcan-api/vulcan-api .

COPY config.toml .
COPY run.sh .

CMD [ "./run.sh" ]
