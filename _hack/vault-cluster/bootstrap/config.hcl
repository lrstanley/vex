ui = true
disable_mlock = true
api_addr = "http://NODE_ID:8200"
cluster_addr = "http://NODE_ID:8201"
max_lease_ttl = "8766h"
default_lease_ttl = "48h"

listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = true
}

storage "raft" {
  path    = "/vault/data"
  node_id = "NODE_ID"
  retry_join {
    leader_api_addr = "http://vault1:8200"
  }
  # retry_join {
  #   leader_api_addr = "http://vault2:8200"
  # }
  # retry_join {
  #   leader_api_addr = "http://vault3:8200"
  # }
}
