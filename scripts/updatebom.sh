#!/usr/bin/env bash

set -e

if ! [[ "$0" =~ scripts/updatebom.sh ]]; then
	echo "must be run from repository root"
	exit 255
fi

echo "installing 'bill-of-materials.json'"
go get -v -u github.com/coreos/license-bill-of-materials

echo "generating bill-of-materials.json"
license-bill-of-materials \
    --override-file ./bill-of-materials.override.json \
    github.com/friendlyhank/etcd-3.4-annotated github.com/friendlyhank/etcd-3.4-annotated/etcdctl > bill-of-materials.json

echo "generated bill-of-materials.json"
