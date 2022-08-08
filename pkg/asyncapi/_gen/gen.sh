#!/bin/bash

# Copyright 2022 Adevinta

set -eu

if [ $# -ne 2 ]; then
	echo "usage: $0 <asyncdoc_path>" >&2
	exit 2
fi

TOOL_DIR="$(realpath "$0")"
TOOL_DIR="$(dirname "$TOOL_DIR")"
# Set working directory to the root of the repo.
cd "$(dirname $0)/../../.."

SOURCE_DIR=$(dirname $(realpath "$1"))
SOURCE_FILE="$(basename $(realpath "$1"))"
SOURCE_FILE="/source/${SOURCE_FILE}"

GO_PACKAGE_NAME="asyncapi"

docker run \
	--rm \
	-v "${TOOL_DIR}:/work" \
	-v "${SOURCE_DIR}:/source" \
	-w "/work" \
	"node:18.3.0-alpine3.15" \
	/bin/sh -c "
    npm install --silent &&
    node gen.js ${SOURCE_FILE} ${GO_PACKAGE_NAME}"
