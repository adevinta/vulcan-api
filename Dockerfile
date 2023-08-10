# syntax=docker/dockerfile:1.4
# Copyright 2021 Adevinta

FROM golang:1.21-alpine3.18 as builder
# Required because the dependency
# https://github.com/confluentinc/confluent-kafka-go requires the gcc compiler.
RUN apk add gcc libc-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH

WORKDIR /app/cmd/vulcan-api
RUN go build -tags musl .

FROM alpine:3.18

WORKDIR /flyway

RUN apk add --no-cache --update openjdk8-jre-base bash gettext libc6-compat

ARG FLYWAY_VERSION=9.19.3

RUN wget -q https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}.tar.gz \
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

COPY --link config.toml run.sh .
COPY --from=builder --link /app/cmd/vulcan-api/vulcan-api .

CMD [ "./run.sh" ]
