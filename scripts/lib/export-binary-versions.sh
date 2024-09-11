#!/bin/bash -eu
LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)" # Figure out where the script is running
PROJECT_DIR="${LIB_DIR}"/../../
export FFI_VERSION=v$(grep "version:     \"" "$PROJECT_DIR"/installer/installer.go | awk -F'"' '{print $2}' )