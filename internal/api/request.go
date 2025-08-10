// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

type wrappedResponse[T any] struct {
	RequestID     string              `json:"request_id"`
	LeaseID       string              `json:"lease_id"`
	LeaseDuration int                 `json:"lease_duration"`
	Renewable     bool                `json:"renewable"`
	Data          T                   `json:"data"`
	Warnings      []string            `json:"warnings"`
	Auth          *api.SecretAuth     `json:"auth,omitempty"`
	WrapInfo      *api.SecretWrapInfo `json:"wrap_info,omitempty"`
	MountType     string              `json:"mount_type,omitempty"`
}

// request is a generic wrapper for HTTP requests to Vault, for requests that may
// not be supported by the Go client.
func request[T any](c *client, method, path string, params map[string]any, data any) (T, error) {
	var v T
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	uri, err := url.Parse(strings.TrimSuffix(c.api.Address(), "/") + path)
	if err != nil {
		return v, err
	}

	if len(params) > 0 {
		query := uri.Query()
		for k, v := range params {
			query.Set(k, fmt.Sprintf("%v", v))
		}
		uri.RawQuery = query.Encode()
	}

	var body io.Reader

	if data != nil {
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(data)
		if err != nil {
			return v, err
		}
		body = &buf
	} else {
		body = http.NoBody
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		uri.String(),
		body,
	)
	if err != nil {
		return v, err
	}

	req.Header = c.api.Headers()
	req.Header.Set("X-Vault-Token", c.api.Token())

	if method != http.MethodGet && method != http.MethodHead && data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errors api.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errors)
		if err == nil && len(errors.Errors) > 0 {
			return v, fmt.Errorf("request failed: %s", strings.Join(errors.Errors, ", "))
		}
		return v, fmt.Errorf("request failed: %s", resp.Status)
	}

	switch resp.Header.Get("Content-Type") {
	case "application/json":
		err = json.NewDecoder(resp.Body).Decode(&v)
		if err != nil {
			return v, fmt.Errorf("request failed: %w", err)
		}
	default:
		slog.Warn("unhandled content type for request", "content-type", resp.Header.Get("Content-Type"))
		var body []byte

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return v, fmt.Errorf("request failed: %w", err)
		}

		// if T is of type []byte, then we can just return the body as is.
		if _, ok := any(v).([]byte); ok {
			v = any(body).(T) //nolint:errcheck
			return v, nil
		}

		if _, ok := any(v).(string); ok {
			v = any(string(body)).(T) //nolint:errcheck
			return v, nil
		}

		return v, fmt.Errorf("unhandled content type for request: %s", resp.Header.Get("Content-Type"))
	}

	return v, nil
}

// ConcurrentLimiter is an HTTP RoundTripper that limits the number of concurrent requests.
// It wraps another RoundTripper and ensures that only a maximum number of requests
// can be processed simultaneously, while allowing unlimited goroutines to queue up.
type ConcurrentLimiter struct {
	// The underlying RoundTripper to delegate requests to.
	Transport http.RoundTripper
	// Semaphore to limit concurrent requests.
	semaphore chan struct{}
}

// NewConcurrentLimiter creates a new ConcurrentLimiter with the specified maximum
// number of concurrent requests. If transport is nil, [http.DefaultTransport] is used.
func NewConcurrentLimiter(maxConcurrent int, transport http.RoundTripper) *ConcurrentLimiter {
	if transport == nil {
		transport = http.DefaultTransport
	}

	return &ConcurrentLimiter{
		Transport: transport,
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// RoundTrip implements [http.RoundTripper] interface. It acquires a semaphore slot
// before making the request and releases it after the request completes.
func (cl *ConcurrentLimiter) RoundTrip(req *http.Request) (*http.Response, error) {
	// Acquire a semaphore slot to limit concurrent requests.
	cl.semaphore <- struct{}{}
	defer func() {
		// Release the semaphore slot when the request completes.
		<-cl.semaphore
	}()

	// Delegate the actual request to the underlying transport.
	return cl.Transport.RoundTrip(req)
}
