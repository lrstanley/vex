// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/types"
)

var mockMounts = []*types.Mount{
	{
		Path: "kv-v1-1/",
		MountOutput: &vapi.MountOutput{
			Description: "KV v1",
			UUID:        "abc-123",
			Type:        "kv",
			Options: map[string]string{
				"version": "1",
			},
			SealWrap:          false,
			PluginVersion:     "v0.24.0+builtin",
			DeprecationStatus: "supported",
		},
	},
	{
		Path: "kv-v2-1/",
		MountOutput: &vapi.MountOutput{
			Description: "KV v2",
			UUID:        "abc-123-2",
			Type:        "kv",
			Options: map[string]string{
				"version": "2",
			},
		},
	},
}

var mockPolicyList = []string{
	"test-policy-1",
	"test-policy-2",
	"test-policy-3",
	"test-policy-4",
	"test-policy-5",
}

var mockPolicy = `# Admin policy for general administration
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
}`

var _ types.Client = &MockClient{}

type MockClient struct {
	ShouldError        bool
	firstHealthChecked atomic.Bool
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) Init() tea.Cmd {
	return m.GetHealth()
}

func (m *MockClient) Update(msg tea.Msg) tea.Cmd {
	vm, ok := msg.(types.ClientMsg)
	if !ok {
		return nil
	}

	var cmds []tea.Cmd

	if vm.Error != nil {
		cmds = append(cmds, types.SendStatus(vm.Error.Error(), types.Error, 2*time.Second))
	} else {
		cmds = append(cmds, types.SendStatus("req success", types.Success, 250*time.Millisecond))
	}

	switch msg := vm.Msg.(type) {
	case types.ClientRequestConfigMsg:
		return m.GetHealth()
	case types.ClientConfigMsg:

		cmds = append(cmds, types.CmdAfterDuration(m.GetHealth(), HealthCheckInterval))

		if !m.firstHealthChecked.Load() && msg.Health != nil {
			m.firstHealthChecked.Store(true)
			cmds = append(cmds, types.SendStatus("initial health checked", types.Success, 1*time.Second))
		}
	}

	return tea.Batch(cmds...)
}

func (m *MockClient) ErrorOr(uuid string, msg tea.Msg) tea.Cmd {
	if m.ShouldError {
		return func() tea.Msg {
			return types.ClientMsg{
				UUID:  uuid,
				Error: errors.New("test error"),
			}
		}
	}
	return func() tea.Msg {
		return types.ClientMsg{
			UUID: uuid,
			Msg:  msg,
		}
	}
}

func (m *MockClient) GetHealth() tea.Cmd {
	return m.ErrorOr("", types.ClientConfigMsg{
		Address: "http://localhost:8200",
		Health: &vapi.HealthResponse{
			Initialized: true,
			Sealed:      false,
			Standby:     false,
			Version:     "1.2.3",
			ClusterName: "test-cluster",
			ClusterID:   "test-cluster-id",
		},
	})
}

func (m *MockClient) ListACLPolicies(uuid string) tea.Cmd {
	return m.ErrorOr(uuid, types.ClientListACLPoliciesMsg{
		Policies: mockPolicyList,
	})
}

func (m *MockClient) GetACLPolicy(uuid, policyName string) tea.Cmd {
	return m.ErrorOr(uuid, types.ClientGetACLPolicyMsg{
		Name:    policyName,
		Content: mockPolicy,
	})
}

func (m *MockClient) GetConfigState(uuid string) tea.Cmd {
	data := `{
    "request_id": "8fd3a29c-ca8a-3909-290b-39b551f65ade",
    "lease_id": "",
    "renewable": false,
    "lease_duration": 0,
    "data": {
        "administrative_namespace_path": "",
        "allow_audit_log_prefixing": false,
        "api_addr": "http://localhost:8200",
        "cache_size": 0,
        "cluster_addr": "http://localhost:8201",
        "storage": {
            "cluster_addr": "http://localhost:8201",
            "disable_clustering": false,
            "raft": {
                "max_entry_size": "",
                "node_id": "localhost",
                "path": "/vault/data",
                "retry_join": "[{\"leader_api_addr\":\"http://localhost:8200\"}]"
            },
            "redirect_addr": "http://localhost:8200",
            "type": "raft"
        }
    },
    "mount_type": "system"
}`

	var out json.RawMessage
	err := json.Unmarshal([]byte(data), &out)
	if err != nil {
		panic(err)
	}
	return m.ErrorOr(uuid, types.ClientConfigStateMsg{Data: out})
}

func (m *MockClient) ListMounts(uuid string, _ ...string) tea.Cmd {
	return m.ErrorOr(uuid, types.ClientListMountsMsg{
		Mounts: mockMounts,
	})
}

func (m *MockClient) ListSecrets(uuid string, mount *types.Mount, path string) tea.Cmd {
	return m.ErrorOr(uuid, types.ClientListSecretsMsg{
		Values: []*types.SecretListRef{
			{Path: fmt.Sprintf("%s%s/foo/", mount.Path, path), Mount: mount},
			{Path: fmt.Sprintf("%s%s/bar", mount.Path, path), Mount: mount},
			{Path: fmt.Sprintf("%s%s/baz", mount.Path, path), Mount: mount},
		},
	})
}

func (m *MockClient) ListAllSecretsRecursive(uuid string) tea.Cmd {
	tree := types.ClientSecretTree{
		{
			Mount: mockMounts[0],
			Path:  "kv-v1-1/foo/",
			Leafs: types.ClientSecretTree{
				{
					Mount: mockMounts[0],
					Path:  "kv-v1-1/foo/bar",
				},
			},
		},
	}
	tree.SetParentOnLeafs(nil)
	return m.ErrorOr(uuid, types.ClientListAllSecretsRecursiveMsg{
		Tree:        tree,
		Requests:    10,
		MaxRequests: MaxRecursiveRequests,
	})
}

func (m *MockClient) GetKVv2Metadata(uuid string, mount *types.Mount, path string) tea.Cmd {
	return m.ErrorOr(uuid, types.ClientGetKVv2MetadataMsg{
		Mount: mount,
		Path:  path,
		Metadata: &vapi.KVMetadata{
			CASRequired:    false,
			CreatedTime:    time.Now().Add(-(24 * time.Hour)),
			UpdatedTime:    time.Now().Add(-(12 * time.Hour)),
			CurrentVersion: 2,
			CustomMetadata: map[string]any{
				"foo": "bar",
			},
			Versions: map[string]vapi.KVVersionMetadata{
				"1": {
					Version:     1,
					CreatedTime: time.Now().Add(-(24 * time.Hour)),
				},
				"2": {
					Version:     2,
					CreatedTime: time.Now().Add(-(12 * time.Hour)),
				},
			},
		},
	})
}

func (m *MockClient) ListKVv2Versions(uuid string, mount *types.Mount, path string) tea.Cmd {
	return m.ErrorOr(uuid, types.ClientListKVv2VersionsMsg{
		Mount: mount,
		Path:  path,
		Versions: []vapi.KVVersionMetadata{
			{Version: 1, CreatedTime: time.Now().Add(-(24 * time.Hour))},
			{Version: 2, CreatedTime: time.Now().Add(-(12 * time.Hour))},
		},
	})
}

func (m *MockClient) GetSecret(uuid string, mount *types.Mount, path string) tea.Cmd {
	if strings.Contains(path, "json") {
		return m.ErrorOr(uuid, types.ClientGetSecretMsg{
			Mount: mount,
			Path:  path,
			Data:  map[string]any{"foo": "bar", "bar": "baz", "inner": map[string]any{"foo": "bar", "bar": "baz"}},
		})
	}
	return m.ErrorOr(uuid, types.ClientGetSecretMsg{
		Mount: mount,
		Path:  path,
		Data:  map[string]any{"foo": "bar", "bar": "baz"},
	})
}
