#!/usr/bin/env bash

set -e

name="${1:-}"
if [ -z "$name" ]; then
  echo "no name given"
  exit 1
fi

base="${2:-unmodified}"

cd $(dirname $0)
cd testdata

if [ -f "$name.base.yaml" ]; then
  echo "testcase $name exists already"
  exit 1
fi

cp $base.base.yaml "$name.base.yaml"
cp $base.diff.yaml "$name.diff.yaml"
cp $base.revision.yaml "$name.revision.yaml"
