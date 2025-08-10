// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package vaultelement

import (
	"fmt"
	"net/url"
	"strings"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/formatter"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const (
	MaxAddrWidth             = 30
	MaxTokenDisplayNameWidth = 15
)

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	Address string
	health  *vapi.HealthResponse
	token   *types.TokenLookupResult

	// Styles.
	addrStyle       lipgloss.Style
	userStyle       lipgloss.Style
	tokenTTLStyle   lipgloss.Style
	clusterOptStyle lipgloss.Style
}

func New(app types.AppState) *Model {
	m := &Model{
		ComponentModel: types.ComponentModel{},
		app:            app,
	}

	m.setStyles()
	return m
}

func (m *Model) setStyles() {
	m.addrStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Background(styles.Theme.StatusBarAddrBg()).
		Foreground(styles.Theme.StatusBarAddrFg())

	m.clusterOptStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Foreground(styles.Theme.InfoFg()).
		Background(styles.Theme.InfoBg())

	m.userStyle = m.clusterOptStyle.
		Foreground(styles.Theme.StatusBarUserFg()).
		Background(styles.Theme.StatusBarUserBg())

	m.tokenTTLStyle = m.clusterOptStyle.
		Foreground(styles.Theme.StatusBarTokenTTLFg()).
		Background(styles.Theme.StatusBarTokenTTLBg())
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case styles.ThemeUpdatedMsg:
		m.setStyles()
	case types.ClientMsg:
		if msg.Error != nil {
			return nil
		}
		switch msg := msg.Msg.(type) {
		case types.ClientConfigMsg:
			m.health = msg.Health
			uri, err := url.Parse(msg.Address)
			if err == nil {
				m.Address = fmt.Sprintf("%s:%s", uri.Hostname(), uri.Port())
			} else {
				m.Address = msg.Address
			}
		case types.ClientTokenLookupSelfMsg:
			m.token = msg.Result
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	var out []string
	add := func(s string) {
		w := ansi.StringWidth(s)
		if styles.W(out...)+w <= m.Width {
			out = append(out, s)
		}
	}

	if m.health != nil && m.health.ClusterName != "" && !strings.HasPrefix(m.health.ClusterName, "vault-cluster-") {
		add(m.addrStyle.Render(formatter.Trunc(m.health.ClusterName, MaxAddrWidth)))
	} else if m.Address != "" {
		add(m.addrStyle.Render(formatter.Trunc(m.Address, MaxAddrWidth)))
	}

	if m.health != nil {
		if !m.health.Initialized {
			fg, bg := styles.Theme.ByStatus(types.Error)
			add(m.clusterOptStyle.
				Foreground(fg).
				Background(bg).
				Render("not init"))
		} else {
			if m.health.Sealed {
				fg, bg := styles.Theme.ByStatus(types.Warning)
				add(m.clusterOptStyle.
					Foreground(fg).
					Background(bg).
					Render("sealed"))
			} else {
				add(m.clusterOptStyle.Render("unsealed"))
			}
		}
		if m.health.Version != "" {
			add(m.clusterOptStyle.Render("v" + m.health.Version))
		}
	}

	if m.token != nil {
		var dn string
		if m.token.DisplayName != "" {
			dn = m.token.DisplayName
		} else if v := strings.Split(m.token.Path, "/"); len(v) > 0 {
			dn = v[len(v)-1]
		}

		if dn != "" {
			add(m.userStyle.Render(formatter.Trunc(dn, MaxTokenDisplayNameWidth)))
		}

		if v := m.token.ExpireTime; !v.IsZero() {
			add(m.tokenTTLStyle.Render("ttl " + formatter.TimeRelative(v, false)))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}
