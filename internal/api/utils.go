// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	tea "github.com/charmbracelet/bubbletea/v2"
	vapi "github.com/hashicorp/vault/api"
	"github.com/lrstanley/vex/internal/types"
)

func secretToList(secret *vapi.Secret) []string {
	if secret == nil || secret.Data == nil {
		return nil
	}
	var keys []string
	if v, ok := secret.Data["keys"]; ok {
		if vv, vok := v.([]any); vok {
			for _, v := range vv {
				if vv, vvok := v.(string); vvok {
					keys = append(keys, vv)
				}
			}
		}
	}
	return keys
}

func wrapHandler[T any](uuid string, f func() (*T, error)) tea.Cmd {
	return func() tea.Msg {
		out, err := f()
		if err != nil {
			v := new(T)
			return types.ClientMsg{
				UUID:  uuid,
				Error: err,
				Msg:   *v,
			}
		}
		return types.ClientMsg{
			UUID: uuid,
			Msg:  *out,
		}
	}
}
