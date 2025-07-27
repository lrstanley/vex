#!/bin/bash

# get directory name of this file.
export BASE_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

# if the file wasn't sourced, error
if [ "$0" = "$BASH_SOURCE" ]; then
	echo "error: this file must be sourced, not executed"
	exit 1
fi

export VAULT_ADDR="http://localhost:8200"
export VAULT_TOKEN="$(cat "${BASE_DIR}/init/root-token")"

echo "set the following VAULT_* env vars:"
env | grep ^VAULT_
