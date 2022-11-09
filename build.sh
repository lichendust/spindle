#!/bin/bash

# dev builds for Linux and macOS

set -e

platform=$(uname)

go build -tags debug -o build_v2/spindle
echo "${platform,,} debug"