#!/bin/bash

# builds for Windows via Windows Subsystem for Linux

set -e

GOOS="windows"
GOARCH="amd64"
CC="/usr/bin/x86_64-w64-mingw32-gcc"
CXX="/usr/bin/x86_64-w64-mingw32-gcc"

if [[ ! -z $1 ]] && [[ $1 == "release" ]]; then
	mkdir -p "build_v2/windows"
	go build -ldflags "-s -w" -trimpath -o build_v2/windows/spindle.exe
	echo "windows release"
	return
fi

go build -tags debug -o build_v2/spindle
echo "windows debug"