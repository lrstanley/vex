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
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/types"
	"github.com/lrstanley/vex/internal/ui/styles"
)

const addrMaxWidth = 30

var _ types.Component = (*Model)(nil) // Ensure we implement the component interface.

type Model struct {
	types.ComponentModel

	// Core state.
	app types.AppState

	// UI state.
	Address string
	health  *vapi.HealthResponse

	// Styles.
	addrStyle       lipgloss.Style
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
		switch msg := msg.Msg.(type) {
		case types.ClientConfigMsg:
			m.health = msg.Health
			uri, err := url.Parse(msg.Address)
			if err == nil {
				m.Address = fmt.Sprintf("%s:%s", uri.Hostname(), uri.Port())
			} else {
				m.Address = msg.Address
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) View() string {
	var addr string
	if m.health != nil && m.health.ClusterName != "" && !strings.HasPrefix(m.health.ClusterName, "vault-cluster-") {
		addr = m.addrStyle.Render(m.health.ClusterName)
	} else if m.Address != "" {
		addr = m.addrStyle.Render(styles.Trunc(m.Address, addrMaxWidth))
	}

	var clusterOpt string
	if m.health != nil {
		switch {
		case !m.health.Initialized:
			fg, bg := styles.Theme.ByStatus(types.Error)
			clusterOpt = m.clusterOptStyle.
				Foreground(fg).
				Background(bg).
				Render("not init")
		case m.health.Sealed:
			fg, bg := styles.Theme.ByStatus(types.Warning)
			clusterOpt = m.clusterOptStyle.
				Foreground(fg).
				Background(bg).
				Render("sealed")
		case m.health.Version != "":
			clusterOpt = m.clusterOptStyle.Render("v" + m.health.Version)
		default:
			clusterOpt = m.clusterOptStyle.Render("unknown")
		}
	}

	return addr + clusterOpt
}
