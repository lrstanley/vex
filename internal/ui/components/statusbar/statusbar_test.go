// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package statusbar

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/api"
	"github.com/lrstanley/vex/internal/config"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/state"
	"github.com/lrstanley/x/charm/steep"
)

// applyTestVaultData sends client responses the statusbar vault element consumes so
// snapshots include address/cluster, seal/version, and token display fields.
func applyTestVaultData(tb testing.TB, tm *steep.Model) {
	tb.Helper()
	for _, msg := range []tea.Msg{
		types.ClientMsg{
			Msg: types.ClientConfigMsg{
				Address: "https://vault.example.com:443",
				Health: &vapi.HealthResponse{
					Initialized: true,
					Sealed:      false,
					ClusterName: "acme-corp",
					Version:     "1.20.0",
				},
			},
		},
		types.ClientMsg{
			Msg: types.ClientTokenLookupSelfMsg{
				Result: &types.TokenLookupResult{
					DisplayName: "dev-admin",
				},
			},
		},
	} {
		tm.Send(msg)
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	t.Run("basic-statusbar", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := steep.NewViewModel(t, m)
		applyTestVaultData(t, tm)
		tm.WaitContainsStrings(t, []string{
			config.AppName,
			"acme-corp",
			"unsealed",
			"v1.20.0",
			"dev-admin",
		})
		tm.RequireSnapshotNoANSI(t)
	})

	t.Run("0-width-height", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(0, 0))
		tm.WaitSettleMessages(t).ExpectDimensions(t, 0, 0)
	})

	t.Run("small-dimensions", func(t *testing.T) {
		t.Parallel()
		app := state.NewMockAppState(api.NewMockClient(), nil)
		m := New(app)
		tm := steep.NewViewModel(t, m, steep.WithInitialTermSize(40, 3))
		applyTestVaultData(t, tm)
		tm.WaitContainsStrings(t, []string{
			config.AppName,
			"acme-corp",
			"unsealed",
		})
		tm.RequireSnapshotNoANSI(t)
	})
}
