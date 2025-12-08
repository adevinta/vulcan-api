# Copyright 2021 Adevinta

FROM --platform=linux/$TARGETARCH golang:1.25-alpine AS builder
# Required because the dependency
# https://github.com/confluentinc/confluent-kafka-go requires the gcc compiler.
RUN apk add --no-cache gcc musl-dev cyrus-sasl-dev mold

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG TARGETOS TARGETARCH

RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH \
    # explicitly link to libsasl2 installed as part of cyrus-sasl-dev
    CGO_LDFLAGS="-fuse-ld=mold -lsasl2" \
    go build -tags musl -ldflags "-w -s" ./cmd/vulcan-api

FROM alpine:3.23

WORKDIR /flyway

RUN apk add --no-cache openjdk17-jre-headless bash gettext cyrus-sasl libgcc

ARG FLYWAY_VERSION=10.10.0

RUN wget -q https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/${FLYWAY_VERSION}/flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && tar -xzf flyway-commandline-${FLYWAY_VERSION}.tar.gz --strip 1 \
    && rm flyway-commandline-${FLYWAY_VERSION}.tar.gz \
    && find ./drivers/ -type f -not -name '*postgres*' -not -name '*jackson*' -delete \
    && chown -R root:root . \
    && ln -s /flyway/flyway /bin/flyway

WORKDIR /app

COPY db/sql /app/sql/

RUN mkdir -p /app/output

COPY --link config.toml run.sh ./
COPY --from=builder --link /app/vulcan-api ./

CMD [ "./run.sh" ]
