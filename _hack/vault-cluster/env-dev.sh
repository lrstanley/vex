#!/bin/bash

# get directory name of this file.
export BASE_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

# if the file wasn't sourced, error
if [ "$0" = "$BASH_SOURCE" ]; then
	echo "error: this file must be sourced, not executed"
	exit 1
fi

USER_ID=${1:?usage: $0 <user-id>}

USERNAME=$(sed -n 's/^username: \(.*\)/\1/p' < "${BASE_DIR}/init/user-credentials/dev${USER_ID}.txt")
PASSWORD=$(sed -n 's/^password: \(.*\)/\1/p' < "${BASE_DIR}/init/user-credentials/dev${USER_ID}.txt")

export VAULT_ADDR="http://localhost:8200"

export VAULT_TOKEN=$(vault login -method=userpass -token-only username="${USERNAME}" password="${PASSWORD}")

echo "set the following VAULT_* env vars:"
env | grep ^VAULT_
