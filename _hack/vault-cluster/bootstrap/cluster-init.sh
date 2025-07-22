#!/bin/bash
# shellcheck disable=SC2155

ROOT_TOKEN=$(cat /init/root-token)
UNSEAL_KEYS=$(cat /init/unseal-keys)

# wait_online <node> <seconds>
wait_online() {
	echo "waiting for ${1} to be online"
	export VAULT_ADDR="http://${1}:8200"

	COUNT=0
	while true; do
		COUNT=$((COUNT + 1))
		nc -z "$1" 8200 && {
			sleep 2
			return 0
		}
		sleep 1
		[ "$COUNT" -ge "$2" ] && return 1
	done
}

# init_node <node>
init_node() {
	echo "initializing ${1}"
	export VAULT_ADDR="http://${1}:8200"

	vault operator init \
		-key-shares=1 \
		-key-threshold=1 \
		-format=json > /tmp/init-data.json || exit 1

	cat /tmp/init-data.json

	jq -r '.root_token' /tmp/init-data.json > /init/root-token || exit 1
	jq -r '.unseal_keys_b64[]' /tmp/init-data.json > /init/unseal-keys || exit 1

	echo "root token: $(cat /init/root-token)"
	echo "unseal keys:"
	cat /init/unseal-keys

	export ROOT_TOKEN=$(cat /init/root-token)
	export UNSEAL_KEYS=$(cat /init/unseal-keys)
}

# is_unsealed <node>
is_unsealed() {
	export VAULT_ADDR="http://${1}:8200"
	vault status > /dev/null 2>&1
	EXIT_CODE="$?"

	if [ "$EXIT_CODE" -eq 0 ]; then
		echo "node ${1} is unsealed"
		return 0
	else
		echo "node ${1} is sealed or other error"
		return 2
	fi
}

# unseal_node <node> <max-seconds>
unseal_node() {
	echo "unsealing ${1}"
	export VAULT_ADDR="http://${1}:8200"

	vault operator unseal -reset > /dev/null 2>&1
	while read -r KEY; do
		echo "unsealing ${1} with key ${KEY}"
		vault operator unseal "$KEY" > /dev/null 2>&1
		sleep 1
	done < /init/unseal-keys

	COUNT=0
	while true; do
		COUNT=$((COUNT + 1))
		is_unsealed "$1" && return 0
		[ "$COUNT" -ge "$2" ] && return 1
		sleep 1
	done
}

# if not initialized, initialize.
if [ -z "$ROOT_TOKEN" ] || [ -z "$UNSEAL_KEYS" ]; then
	wait_online "vault1" 20
	init_node vault1 || exit 1
	sleep 3
fi

for NODE in vault1 vault2 vault3; do
	# if sealed, unseal.
	is_unsealed "$NODE" || {
		COUNT=0
		while true; do
			COUNT=$((COUNT + 1))
			unseal_node "$NODE" 30 && break
			[ "$COUNT" -ge 5 ] && exit 1
			sleep 1
		done
	}
done

export VAULT_ADDR="http://vault1:8200"
export VAULT_TOKEN="$ROOT_TOKEN"

while true; do
	echo "waiting for vault1 to respond to status calls"
	vault status > /dev/null 2>&1 && break
	sleep 1
done

/generate-data.sh
