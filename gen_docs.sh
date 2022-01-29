#!/bin/bash

set -euxo pipefail

GOTOOL="$GOROOT/bin/go"

rm -r docs/
mkdir -p docs/
echo -n "" > docs/index.md

pkgs="$($GOTOOL list ./... | sed "s/^github.com\/bradenaw\/juniper\///g" | grep -v "internal")"

while read pkg; do
    mkdir -p "$(dirname "docs/$pkg")"
    echo "[$pkg]($pkg.md)" >> docs/index.md
    echo >> docs/index.md
done <<< "$pkgs"

xargs -I{} -P12 --verbose bash -c $GOTOOL' run github.com/bradenaw/juniper/internal/gendocs {} > "docs/{}.md"' <<< "$pkgs"