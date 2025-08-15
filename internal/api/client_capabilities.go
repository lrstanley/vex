// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"net/http"

	"github.com/lrstanley/vex/internal/types"
)

func (c *client) getCapabilities(paths ...string) (map[string]types.ClientCapabilities, error) {
	// Current Sys.Capabilities* vault methods only check 1 path.
	// TODO: https://github.com/hashicorp/vault/issues/31376

	if len(paths) == 0 {
		return map[string]types.ClientCapabilities{}, nil
	}

	results, err := request[*wrappedResponse[map[string]types.ClientCapabilities]](
		c,
		http.MethodPost,
		"/v1/sys/capabilities-self",
		nil,
		map[string]any{"paths": paths},
	)
	if err != nil {
		return nil, err
	}

	delete(results.Data, "capabilities") // API compatibility, don't need this.
	return results.Data, nil
}
