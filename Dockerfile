# syntax=docker/dockerfile:1.4
# Copyright 2021 Adevinta

FROM --platform=linux/$TARGETARCH golang:1.22-alpine3.19 as builder
# Required because the dependency
# https://github.com/confluentinc/confluent-kafka-go requires the gcc compiler.
RUN apk add --no-cache gcc musl-dev cyrus-sasl-dev mold

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH
RUN echo $TARGETARCH

WORKDIR /app/cmd/vulcan-api
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH \
    # explicitly link to libsasl2 installed as part of cyrus-sasl-dev
    CGO_LDFLAGS="-fuse-ld=mold -lsasl2" \
    go build -tags musl -ldflags "-w -s" .

FROM alpine:3.20

WORKDIR /flyway

RUN apk add --no-cache --update openjdk17-jre-headless bash gettext cyrus-sasl

ARG FLYWAY_VERSION=10.10.0

RUN wget -q https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && tar -xzf flyway-commandline-${FLYWAY_VERSION}.tar.gz --strip 1 \
    && rm flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && find ./drivers/ -type f | grep -Ev '(postgres|jackson)' | xargs rm \
    && chown -R root:root . \
    && ln -s /flyway/flyway /bin/flyway

ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"

ENV BUILD_RFC3339 "$BUILD_RFC3339"
ENV COMMIT "$COMMIT"

WORKDIR /app

COPY db/sql /app/sql/

RUN mkdir -p /app/output

COPY --link config.toml run.sh ./
COPY --from=builder --link /app/cmd/vulcan-api/vulcan-api ./

CMD [ "./run.sh" ]
