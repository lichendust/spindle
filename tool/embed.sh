#!/bin/bash

# hard-wraps the "help" command text into the
# codebase to ensure consistency

set -e

version=$(grep 'const VERSION' source/*.go | awk -F"[ \"]+" '/VERSION/{print $4}')

echo -n "VERSION = $version" > manual/config/templates/version.x

pushd manual > /dev/null
spindle.exe build
popd > /dev/null

target=source/data_manual.go

cat tool/header_license.txt > $target
printf "

// this file was generated by tool/embed.sh: don't modify!

package main

func manual_content(arg string) string {
	switch arg {
" >> $target

for f in manual/public/*; do
	name=$(basename ${f%.html})
	name=${name/*_}

	if [ $name == "index" ]; then
		printf "\tdefault:\n\t\treturn \`" >> $target
	else
		printf "\tcase \"$name\":\n\t\treturn \`" >> $target
	fi

	echo -n "$(cat $f)" >> $target
	printf "\`\n" >> $target
done

printf "\t}\n}" >> $target