#!/bin/bash

set -e

source=_parser_string.go
target=parser_string.go

pushd source > /dev/null

stringer.exe -type=AST_Type,File_Type,Path_Type,Exec_Type,Modifier -output=$source

cat ../text/header_license.txt > $target
printf "

//go:build debug

" >> $target
cat $source >> $target

rm $source

popd > /dev/null