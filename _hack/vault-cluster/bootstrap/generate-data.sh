#!/bin/bash
# shellcheck disable=SC2155

export EXISTING_ENGINES=$(vault secrets list -format=json | jq -r '. | keys[]')

# KVv1
for i in $(seq 1 10); do
	if ! grep -q "kv-v1-${i}" <<< "$EXISTING_ENGINES"; then
		echo "enabling kv-v1-${i}"
		vault secrets enable \
			-path="kv-v1-${i}" \
			-description="kv v1 secret engine created with vault-cluster-init" \
			-version=1 kv > /dev/null || exit 1
	fi

	for j in $(seq 1 10); do
		echo "writing kv-v1-${i}/secret-${j}"
		vault kv put "kv-v1-${i}/secret-${j}" \
			key1="value1-${j}" \
			key2="value2-${j}" \
			key3="value3-${j}" > /dev/null || exit 1
	done
done

# KVv2
for i in $(seq 1 10); do
	if ! grep -q "kv-v2-${i}" <<< "$EXISTING_ENGINES"; then
		echo "enabling kv-v2-${i}"
		vault secrets enable \
			-path="kv-v2-${i}" \
			-description="kv v2 secret engine created with vault-cluster-init" \
			-version=2 kv > /dev/null || exit 1
	fi

	for j in $(seq 1 10); do
		echo "writing kv-v2-${i}/secret-${j}"
		vault kv put "kv-v2-${i}/secret-${j}" \
			key1="value1-${j}-$(date +%s)" \
			key2="value2-${j}-$(date +%s)" \
			key3="value3-${j}-$(date +%s)" > /dev/null || exit 1

		echo "writing kv-v2-${i}/secret-${j} metadata"
		vault kv metadata put \
			-mount="kv-v2-${i}" \
			-custom-metadata="foo=bar" \
			-custom-metadata="baz=qux" \
			-custom-metadata="written-at=$(date +%s)" \
			"secret-${j}" > /dev/null || exit 1
	done
done

# transit
if ! grep -q "transit" <<< "$EXISTING_ENGINES"; then
	echo "enabling transit"
	vault secrets enable transit > /dev/null || /bin/true
fi

# enable userpass auth method
echo "enabling userpass auth method"
vault auth enable userpass > /dev/null || /bin/true

# create policies
echo "creating policies"

# root policy with sudo permissions
cat > /tmp/root-policy.hcl << 'EOF'
# Root policy with sudo permissions
path "*" {
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
EOF
echo "writing root-policy"
vault policy write root-policy /tmp/root-policy.hcl > /dev/null || exit 1

# admin policy for general administration
cat > /tmp/admin-policy.hcl << 'EOF'
# Admin policy for general administration
path "auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/auth/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "sys/policies/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv-v1-*/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv-v2-*/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
echo "writing admin-policy"
vault policy write admin-policy /tmp/admin-policy.hcl > /dev/null || exit 1

# user policy for basic access
cat > /tmp/user-policy.hcl << 'EOF'
# User policy for basic access
path "kv-v1-1/*" {
  capabilities = ["read", "list"]
}

path "kv-v2-1/*" {
  capabilities = ["read", "list"]
}
EOF
echo "writing user-policy"
vault policy write user-policy /tmp/user-policy.hcl > /dev/null || exit 1

# readonly policy
cat > /tmp/readonly-policy.hcl << 'EOF'
# Readonly policy
path "kv-v1-1/*" {
  capabilities = ["read"]
}

path "kv-v2-1/*" {
  capabilities = ["read"]
}
EOF
echo "writing readonly-policy"
vault policy write readonly-policy /tmp/readonly-policy.hcl > /dev/null || exit 1

# create users
echo "creating users"

# root user with sudo permissions
echo "writing root user"
vault write auth/userpass/users/root \
	password="rootpass123" \
	policies="root-policy" > /dev/null || exit 1

# admin user
echo "writing admin user"
vault write auth/userpass/users/admin \
	password="adminpass123" \
	policies="admin-policy" > /dev/null || exit 1

# regular users
for i in $(seq 1 5); do
	echo "writing user${i}"
	vault write auth/userpass/users/user${i} \
		password="userpass${i}" \
		policies="user-policy" > /dev/null || exit 1
done

# readonly users
for i in $(seq 1 3); do
	echo "writing readonly${i}"
	vault write auth/userpass/users/readonly${i} \
		password="readonlypass${i}" \
		policies="readonly-policy" > /dev/null || exit 1
done

# write root user credentials to accessible mount
echo "writing root user credentials to accessible mount"
mkdir -p /init/user-credentials
cat > /init/user-credentials/root-user.txt << EOF
username: root
password: rootpass123
policies: root-policy
description: root user with sudo permissions
EOF

cat > /init/user-credentials/admin-user.txt << EOF
username: admin
password: adminpass123
policies: admin-policy
description: admin user with general administration permissions
EOF

# create additional mock users with random policies
echo "creating additional mock users"

# developer users
for i in $(seq 1 3); do
	cat > "/tmp/dev-policy-${i}.hcl" << EOF
# Developer policy ${i}
path "kv-v1-${i}/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "kv-v2-${i}/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF
	echo "writing dev-policy-${i}"
	vault policy write "dev-policy-${i}" "/tmp/dev-policy-${i}.hcl" || exit 1

	echo "writing dev${i}"
	vault write "auth/userpass/users/dev${i}" \
		password="devpass${i}" \
		policies="dev-policy-${i}" || exit 1

	echo "writing dev${i} credentials"
	cat > "/init/user-credentials/dev${i}.txt" << EOF
username: dev${i}
password: devpass${i}
policies: dev-policy-${i}
description: developer user with ${i} permissions
EOF
done

# ops users
for i in $(seq 1 2); do
	cat > "/tmp/ops-policy-${i}.hcl" << EOF
# Operations policy ${i}
path "sys/health" {
  capabilities = ["read"]
}

path "kv-v1-*/*" {
  capabilities = ["read", "list"]
}

path "kv-v2-*/*" {
  capabilities = ["read", "list"]
}
EOF
	echo "writing ops-policy-${i}"
	vault policy write ops-policy-${i} /tmp/ops-policy-${i}.hcl || exit 1

	echo "writing ops${i}"
	vault write auth/userpass/users/ops${i} \
		password="opspass${i}" \
		policies="ops-policy-${i}" || exit 1

	echo "writing ops${i} credentials"
	cat > "/init/user-credentials/ops${i}.txt" << EOF
username: ops${i}
password: opspass${i}
policies: ops-policy-${i}
description: operations user with ${i} permissions
EOF
done

echo "userpass auth setup complete"
echo "root user credentials written to /init/user-credentials/"
echo "total users created:"
echo "  - root (sudo permissions)"
echo "  - admin (admin permissions)"
echo "  - user1-5 (basic user permissions)"
echo "  - readonly1-3 (readonly permissions)"
echo "  - dev1-3 (developer permissions)"
echo "  - ops1-2 (operations permissions)"
