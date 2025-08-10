// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"net/http"

	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/lrstanley/vex/internal/types"
)

func (c *client) TokenLookupSelf(uuid string) tea.Cmd {
	return wrapHandler(uuid, func() (*types.ClientTokenLookupSelfMsg, error) {
		data, err := request[*wrappedResponse[types.TokenLookupResult]](
			c,
			http.MethodGet,
			"/v1/auth/token/lookup-self",
			nil,
			nil,
		)
		if err != nil {
			return nil, err
		}
		return &types.ClientTokenLookupSelfMsg{Result: &data.Data}, nil
	})
}
