#!/usr/bin/env bash
set -e
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
CACHE_DIR="docker_images"
mkdir -p $CACHE_DIR
echo "pulling cached docker img"
docker load -i $CACHE_DIR/images.tar || true
docker pull astraone/localnet-test
echo "saving cached docker img"
docker save -o $CACHE_DIR/images.tar astraone/localnet-test
docker run -v "$DIR/../:/go/src/github.com/Astra-Net/AstraNetwork" astraone/localnet-test -n