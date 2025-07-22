#!/usr/bin/env sh

set -ex

mkdir -vp /vault/config
cp -av /config.hcl /vault/config/config.hcl
sed -ri "s/NODE_ID/$(hostname)/" /vault/config/config.hcl

chmod 777 /vault/data

/usr/local/bin/docker-entrypoint.sh server
