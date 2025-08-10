// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package types

import (
	"encoding/json"
	"iter"
	"slices"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
)

type Client interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	GetHealth(uuid string) tea.Cmd
	TokenLookupSelf(uuid string) tea.Cmd
	ListMounts(uuid string, filterTypes ...string) tea.Cmd
	ListSecrets(uuid string, mount *Mount, path string) tea.Cmd
	ListAllSecretsRecursive(uuid string) tea.Cmd
	GetSecret(uuid string, mount *Mount, path string) tea.Cmd
	GetKVv2Metadata(uuid string, mount *Mount, path string) tea.Cmd
	ListACLPolicies(uuid string) tea.Cmd
	GetACLPolicy(uuid string, policyName string) tea.Cmd
	GetConfigState(uuid string) tea.Cmd
}

// ClientMsg is a wrapper for any message relating to Vault API/client/etc responses.
type ClientMsg struct {
	UUID  string `json:"uuid"`
	Msg   any    `json:"msg,omitempty"`
	Error error  `json:"error,omitempty"`
}

// ClientRequestConfigMsg is a message to request the Vault configuration.
type ClientRequestConfigMsg struct{}

// ClientConfigMsg is a message containing the Vault configuration. Health field is
// nil if the last health check failed.
type ClientConfigMsg struct {
	Address string               `json:"address"`
	Health  *vapi.HealthResponse `json:"health,omitempty"`
}

type ClientListMountsMsg struct {
	Mounts []*Mount `json:"mounts"`
}

type Mount struct {
	*vapi.MountOutput
	Path string `json:"path"`
}

func (m *Mount) KVVersion() int {
	ver, ok := m.Options["version"]
	if !ok {
		return -1
	}
	v, _ := strconv.Atoi(ver)
	if v < 1 {
		return -1
	}
	return v
}

type ClientListSecretsMsg struct {
	Values []*SecretListRef `json:"values"`
}

type SecretListRef struct {
	Mount *Mount `json:"mount"`
	Path  string `json:"path"`
}

type ClientGetSecretMsg struct {
	Mount *Mount         `json:"mount"`
	Path  string         `json:"path"`
	Data  map[string]any `json:"data"`
}

type ClientGetKVv2MetadataMsg struct {
	Mount    *Mount           `json:"mount"`
	Path     string           `json:"path"`
	Metadata *vapi.KVMetadata `json:"metadata"`
}

type ClientListKVv2VersionsMsg struct {
	Mount    *Mount                   `json:"mount"`
	Path     string                   `json:"path"`
	Versions []vapi.KVVersionMetadata `json:"versions"`
}

type ClientListACLPoliciesMsg struct {
	Policies []string `json:"policies"`
}

type ClientGetACLPolicyMsg struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type ClientConfigStateMsg struct {
	Data json.RawMessage `json:"data"`
}

type ClientCapability string

var (
	CapabilityRead   ClientCapability = "read"
	CapabilityWrite  ClientCapability = "write"
	CapabilityDelete ClientCapability = "delete"
	CapabilityList   ClientCapability = "list"
	CapabilityDeny   ClientCapability = "deny"
	CapabilityRoot   ClientCapability = "root"
	CapabilitySudo   ClientCapability = "sudo"
)

type ClientCapabilities []ClientCapability

func (c ClientCapabilities) Contains(capability ClientCapability) bool {
	if capability != CapabilityDeny && (slices.Contains(c, CapabilityRoot) || slices.Contains(c, CapabilitySudo)) {
		return true
	}
	return slices.Contains(c, capability)
}

func (c ClientCapabilities) String() string {
	values := make([]string, 0, len(c))
	for _, v := range c {
		values = append(values, string(v))
	}
	return strings.Join(values, ", ")
}

type ClientSecretTree []*ClientSecretTreeRef

// IterRefs iterates over all the leaf refs in the tree, recursively.
func (c ClientSecretTree) IterRefs() iter.Seq[*ClientSecretTreeRef] {
	return func(yield func(*ClientSecretTreeRef) bool) {
		var iter func(*ClientSecretTreeRef) bool
		iter = func(ref *ClientSecretTreeRef) bool {
			if !yield(ref) {
				return false
			}
			for _, leaf := range ref.Leafs {
				if !iter(leaf) {
					return false
				}
			}
			return true
		}
		for _, ref := range c {
			if !iter(ref) {
				return
			}
		}
	}
}

// SetParentOnLeafs sets the parent on all leafs in the tree.
func (c ClientSecretTree) SetParentOnLeafs(parent *ClientSecretTreeRef) {
	for _, ref := range c {
		if ref.Parent != nil {
			continue
		}
		ref.Parent = parent
		if ref.HasLeafs() {
			ref.Leafs.SetParentOnLeafs(ref)
		}
	}
}

type ClientSecretTreeRef struct {
	Parent *ClientSecretTreeRef `json:"-"`
	Mount  *Mount               `json:"mount"`

	Path         string             `json:"path"`
	Leafs        ClientSecretTree   `json:"leafs,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities,omitempty"`
	Incomplete   bool               `json:"incomplete"`
}

func (c *ClientSecretTreeRef) IsFolder() bool {
	return strings.HasSuffix(c.Path, "/")
}

func (c *ClientSecretTreeRef) IsSecret() bool {
	return !c.IsFolder()
}

func (c *ClientSecretTreeRef) HasLeafs() bool {
	return len(c.Leafs) > 0
}

// GetFullPath returns the full path of the secret, combining all parent paths.
func (c *ClientSecretTreeRef) GetFullPath() string {
	if c.Parent == nil {
		return c.Path
	}
	parent := c.Parent.GetFullPath()
	if parent == "/" {
		parent = ""
	}
	return parent + c.Path
}

// IterRefs iterates over all the leaf refs in this leaf, recursively.
func (c *ClientSecretTreeRef) IterRefs() iter.Seq[*ClientSecretTreeRef] {
	return func(yield func(*ClientSecretTreeRef) bool) {
		var iter func(*ClientSecretTreeRef) bool
		iter = func(ref *ClientSecretTreeRef) bool {
			if !yield(ref) {
				return false
			}
			for _, leaf := range ref.Leafs {
				if !iter(leaf) {
					return false
				}
			}
			return true
		}
		iter(c)
	}
}

func (c *ClientSecretTreeRef) ApplyCapabilities(path string, caps ClientCapabilities) bool {
	if c.GetFullPath() == path {
		c.Capabilities = caps
		return true
	}
	return false
}

type ClientListAllSecretsRecursiveMsg struct {
	Tree        ClientSecretTree `json:"tree"`
	Requests    int64            `json:"requests"`
	MaxRequests int64            `json:"max_requests"`
}

type ClientTokenLookupSelfMsg struct {
	Result *TokenLookupResult `json:"result"`
}

type TokenLookupResult struct {
	Accessor       string         `json:"accessor,omitempty"`
	CreationTime   int64          `json:"creation_time,omitempty"`
	CreationTTL    int64          `json:"creation_ttl,omitempty"`
	DisplayName    string         `json:"display_name,omitempty"`
	EntityID       string         `json:"entity_id,omitempty"`
	ExpireTime     time.Time      `json:"expire_time,omitempty"`
	ExplicitMaxTTL int64          `json:"explicit_max_ttl,omitempty"`
	ID             string         `json:"id,omitempty"`
	IssueTime      time.Time      `json:"issue_time,omitempty"`
	Meta           map[string]any `json:"meta,omitempty"`
	NumUses        int64          `json:"num_uses,omitempty"`
	Orphan         bool           `json:"orphan,omitempty"`
	Path           string         `json:"path,omitempty"`
	Policies       []string       `json:"policies,omitempty"`
	Renewable      bool           `json:"renewable,omitempty"`
	TTL            int64          `json:"ttl,omitempty"`
	Type           string         `json:"type,omitempty"`
}

func (r *TokenLookupResult) WhenExpires() time.Duration {
	return time.Until(r.ExpireTime)
}
