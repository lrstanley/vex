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

	tea "charm.land/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
)

// Client is an interface for interacting with a Vault server.
type Client interface {
	Init() tea.Cmd
	Update(msg tea.Msg) tea.Cmd
	// GetHealth returns a command to get the health of the Vault server.
	GetHealth(uuid string) tea.Cmd
	// TokenLookupSelf returns a command to lookup the current token.
	TokenLookupSelf(uuid string) tea.Cmd

	// ListMounts returns a command to list the mounts of the Vault server.
	ListMounts(uuid string) tea.Cmd
	// ListSecrets returns a command to list the secrets of the Vault server,
	// under a given mount and path. If path is empty, it will list all secrets
	// under the mount.
	ListSecrets(uuid string, mount *Mount, path string) tea.Cmd
	// ListAllSecretsRecursive returns a command to list all secrets of the Vault
	// server. If mount is nil, it will list all secrets under all mounts.
	ListAllSecretsRecursive(uuid string, mount *Mount) tea.Cmd
	// ListKVv2Versions returns a command to list the versions of a KVv2 secret
	// under a given mount and path.
	ListKVv2Versions(uuid string, mount *Mount, path string) tea.Cmd

	// GetKVSecret returns a command to get a secret from the Vault server,
	// under a given mount and path.
	GetKVSecret(uuid string, mount *Mount, path string, version int) tea.Cmd
	// GetKVv2Metadata returns a command to get the metadata of a KVv2 secret
	// under a given mount and path.
	GetKVv2Metadata(uuid string, mount *Mount, path string) tea.Cmd

	// DeleteKVSecret deletes a secret, under a given mount and path.
	DeleteKVSecret(uuid string, mount *Mount, path string, versions ...int) tea.Cmd
	// UndeleteKVSecret undeletes a secret, under a given mount and path.
	UndeleteKVSecret(uuid string, mount *Mount, path string, versions ...int) tea.Cmd
	// DestroyKVSecret destroys a secret, under a given mount and path.
	DestroyKVSecret(uuid string, mount *Mount, path string, versions ...int) tea.Cmd

	// ListACLPolicies returns a command to list the ACL policies of the Vault
	// server.
	ListACLPolicies(uuid string) tea.Cmd
	// GetACLPolicy returns a command to get an ACL policy of the Vault server.
	GetACLPolicy(uuid string, policyName string) tea.Cmd
	// GetConfigState returns a command to get the configuration of the Vault
	// server.
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

// ClientListMountsMsg is a message containing the list of mounts of the Vault
// server.
type ClientListMountsMsg struct {
	Mounts []*Mount `json:"mounts"`
}

// Mount is a mount of the Vault server.
type Mount struct {
	*vapi.MountOutput `json:",inline"`

	// Path is the path of the mount.
	Path string `json:"path"`

	// Capabilities are the capabilities of the mount. Not always available,
	// depends on the query made.
	Capabilities ClientCapabilities `json:"capabilities"`
}

// KVVersion returns the version of the KV mount, if the mount is a KV mount,
// returning -1 otherwise.
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

// PrefixPaths prefixes the given paths with the mount path.
func (m *Mount) PrefixPaths(paths ...string) (prefixed []string) {
	for _, path := range paths {
		prefixed = append(prefixed, m.Path+path)
	}
	return prefixed
}

// ClientListSecretsMsg is a message containing the list of secrets, under a given
// mount and path.
type ClientListSecretsMsg struct {
	Values []*SecretListRef `json:"values"`
}

// SecretListRef is a reference to a secret.
type SecretListRef struct {
	Mount        *Mount             `json:"mount"`
	Path         string             `json:"path"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

// FullPath returns the full path of the secret (including the mount path).
func (r *SecretListRef) FullPath() string {
	return r.Mount.Path + r.Path
}

// ClientGetSecretMsg is a message containing the data of a secret, under a given
// mount and path.
type ClientGetSecretMsg struct {
	Mount *Mount         `json:"mount"`
	Path  string         `json:"path"`
	Data  map[string]any `json:"data"`
}

// ClientGetKVv2MetadataMsg is a message containing the metadata of a KVv2, under
// a given mount and path.
type ClientGetKVv2MetadataMsg struct {
	Mount    *Mount           `json:"mount"`
	Path     string           `json:"path"`
	Metadata *vapi.KVMetadata `json:"metadata"`
}

// ClientListKVv2VersionsMsg is a message containing the versions of a KVv2
// secret, under a given mount and path.
type ClientListKVv2VersionsMsg struct {
	Mount    *Mount                   `json:"mount"`
	Path     string                   `json:"path"`
	Versions []vapi.KVVersionMetadata `json:"versions"`
}

// ClientListACLPoliciesMsg is a message containing a list of all ACL policies.
type ClientListACLPoliciesMsg struct {
	Policies []string `json:"policies"`
}

// ClientGetACLPolicyMsg is a message containing the data of an ACL policy.
type ClientGetACLPolicyMsg struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ClientConfigStateMsg is a message containing the configuration state of the
// cluster.
type ClientConfigStateMsg struct {
	Data json.RawMessage `json:"data"`
}

// ClientCapability contains the capabilities of a given identity, meant to
// determine the level of permissions for a given mount/path/etc.
type ClientCapability string

var (
	// "Note: Capabilities usually map to the HTTP verb, and not the underlying
	// action taken. This can be a common source of confusion. Generating database
	// credentials creates database credentials, but the HTTP request is a GET
	// which corresponds to a read capability. Thus, to grant access to generate
	// database credentials, the policy would grant read access on the appropriate
	// path."
	//  - https://developer.hashicorp.com/vault/docs/concepts/policies

	CapabilityRoot      ClientCapability = "root"
	CapabilityDeny      ClientCapability = "deny"
	CapabilitySudo      ClientCapability = "sudo"
	CapabilityWrite     ClientCapability = "write"
	CapabilitySubscribe ClientCapability = "subscribe"
	CapabilityRecover   ClientCapability = "recover"

	// Capabilities that also have associated HTTP-method mappings.

	CapabilityRead   ClientCapability = "read"   // GET.
	CapabilityList   ClientCapability = "list"   // LIST.
	CapabilityDelete ClientCapability = "delete" // DELETE.
	CapabilityCreate ClientCapability = "create" // POST/PUT.
	CapabilityUpdate ClientCapability = "update" // POST/PUT.
	CapabilityPatch  ClientCapability = "patch"  // PATCH.
)

type ClientCapabilities []ClientCapability

// Contains returns true if the capabilities contain the given capability.
// Accounts for root, sudo, and deny and their associated precedence.
func (c ClientCapabilities) Contains(capability ClientCapability) bool {
	if slices.Contains(c, CapabilityRoot) {
		return true
	}

	// Deny takes precedence over all other capabilities, including sudo.
	if slices.Contains(c, CapabilityDeny) {
		return false
	}

	if slices.Contains(c, CapabilitySudo) {
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
		var iterator func(*ClientSecretTreeRef) bool
		iterator = func(ref *ClientSecretTreeRef) bool {
			if !yield(ref) {
				return false
			}
			for _, leaf := range ref.Leafs {
				if !iterator(leaf) {
					return false
				}
			}
			return true
		}
		for _, ref := range c {
			if !iterator(ref) {
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

// ClientSecretTreeRef is a reference to a path.
type ClientSecretTreeRef struct {
	Parent *ClientSecretTreeRef `json:"-"`
	Mount  *Mount               `json:"mount"`

	Path         string             `json:"path"`
	Leafs        ClientSecretTree   `json:"leafs,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities,omitempty"`
	Incomplete   bool               `json:"incomplete"`
}

// ClientSecretListRef returns a SecretListRef for the secret tree ref.
func (c *ClientSecretTreeRef) ClientSecretListRef() *SecretListRef {
	return &SecretListRef{
		Mount:        c.Mount,
		Path:         c.Path,
		Capabilities: c.Capabilities,
	}
}

// IsFolder returns true if the path is a folder.
func (c *ClientSecretTreeRef) IsFolder() bool {
	return strings.HasSuffix(c.Path, "/")
}

// IsSecret returns true if the path is a secret.
func (c *ClientSecretTreeRef) IsSecret() bool {
	return !c.IsFolder()
}

// HasLeafs returns true if the path has any leafs (children).
func (c *ClientSecretTreeRef) HasLeafs() bool {
	return len(c.Leafs) > 0
}

// GetFullPath returns the full path of the secret, combining all parent paths.
// If withMount is false, the mount path will not be included.
func (c *ClientSecretTreeRef) GetFullPath(withMount bool) string {
	var prefix string
	if !withMount && c.Parent == nil {
		return ""
	}
	if c.Parent != nil {
		prefix = c.Parent.GetFullPath(withMount)
	}
	return prefix + c.Path
}

// IterRefs iterates over all the leaf refs in this leaf, recursively.
func (c *ClientSecretTreeRef) IterRefs() iter.Seq[*ClientSecretTreeRef] {
	return func(yield func(*ClientSecretTreeRef) bool) {
		var iterator func(*ClientSecretTreeRef) bool
		iterator = func(ref *ClientSecretTreeRef) bool {
			if !yield(ref) {
				return false
			}
			for _, leaf := range ref.Leafs {
				if !iterator(leaf) {
					return false
				}
			}
			return true
		}
		iterator(c)
	}
}

// ApplyCapabilities applies the given capabilities to the secret tree ref assuming
// the path matches.
func (c *ClientSecretTreeRef) ApplyCapabilities(path string, caps ClientCapabilities) bool {
	if c.GetFullPath(true) == path {
		c.Capabilities = caps
		return true
	}
	return false
}

// ClientListAllSecretsRecursiveMsg is a message containing the results of a
// recursive list of a given mount/path/etc.
type ClientListAllSecretsRecursiveMsg struct {
	Tree            ClientSecretTree `json:"tree"`
	RequestAttempts int64            `json:"request_attempts"`
	Requests        int64            `json:"requests"`
	MaxRequests     int64            `json:"max_requests"`
}

// ClientTokenLookupSelfMsg is a message containing the result of a token lookup.
type ClientTokenLookupSelfMsg struct {
	Result *TokenLookupResult `json:"result"`
}

// TokenLookupResult is the result of a token lookup.
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

// WhenExpires returns the duration until the token expires.
func (r *TokenLookupResult) WhenExpires() time.Duration {
	return time.Until(r.ExpireTime)
}

type ClientSuccessMsg struct{}
