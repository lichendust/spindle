@echo off

rem dev builds for Windows

go build -tags debug -o build_v2/spindle.exe
echo windows debug