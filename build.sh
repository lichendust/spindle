#!/bin/bash

# builds for Linux and macOS

set -e

platform=$(uname)

if [[ ! -z $1 ]] && [[ $1 == "release" ]]; then
	mkdir -p "build_v2/${platform,,}"
	go build -ldflags "-s -w" -trimpath -o build_v2/${platform,,}/spindle
	echo "${platform,,} release"
	return
fi

go build -tags debug -o build_v2/spindle
echo "${platform,,} debug"