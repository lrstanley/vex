// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

func (c *client) ListACLPolicies(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientListACLPoliciesMsg, error) {
		policies, err := c.api.Sys().ListPolicies()
		if err != nil {
			return nil, fmt.Errorf("list acl policies: %w", err)
		}

		slices.Sort(policies)

		return &types.ClientListACLPoliciesMsg{Policies: policies}, nil
	})
}

func (c *client) GetACLPolicy(uuid string, policyName string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientGetACLPolicyMsg, error) {
		policy, err := c.api.Sys().GetPolicy(policyName)
		if err != nil {
			return nil, fmt.Errorf("get acl policy %s: %w", policyName, err)
		}

		return &types.ClientGetACLPolicyMsg{
			Name:    policyName,
			Content: policy,
		}, nil
	})
}
