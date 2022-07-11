#!/bin/bash

set -eu
if [ $# -ne 2 ]; then
	echo "usage: $0 <asyncdoc_path> <outputdir>" >&2
	exit 2
fi

TOOL_DIR="$(realpath "$0")"
TOOL_DIR="$(dirname "$TOOL_DIR")"
# Set working directory to the root of the repo.
cd "$(dirname $0)/../../.."

# GENDIR="$(realpath pkg/asyncapi/gen)"
# SOURCEDIR="$(realpath docs/async)"
# GENDIR="$(realpath pkg/asyncapi/gen)"

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
    node gen.js ${SOURCE_FILE} ${GO_PACKAGE_NAME}
"
