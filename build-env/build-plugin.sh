#!/bin/bash

set -e
target=$1

function usage() {
  cat <<EOF
To build a plugin:
  $0 <target>
EOF
}

if [ -z "$target" ]; then
  usage
  exit 1
fi

cd /plugin && go build -ldflags '-s -w' -buildmode=plugin -o $target .
