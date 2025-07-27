// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type HTTPRoundTripper struct {
	RoundTripper http.RoundTripper
}

func (rt *HTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.RoundTripper == nil {
		rt.RoundTripper = http.DefaultTransport
	}

	l := slog.Default()

	// Only log if the context has a debug level, otherwise breakout early to avoid
	// overhead.
	if !l.Enabled(req.Context(), slog.LevelDebug) {
		return rt.RoundTripper.RoundTrip(req)
	}

	var pcs [1]uintptr
	_ = runtime.Callers(6, pcs[:]) // Skip this, and all of the net/http/client functions.

	r := slog.NewRecord(time.Now(), slog.LevelDebug, "http request", pcs[0])

	r.AddAttrs(
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("user-agent", req.UserAgent()),
		slog.Int64("content-length", req.ContentLength),
	)

	_ = l.Handler().Handle(req.Context(), r)

	started := time.Now()
	resp, err := rt.RoundTripper.RoundTrip(req)
	duration := time.Since(started)
	if err != nil {
		r = slog.NewRecord(time.Now(), slog.LevelError, "http request failed", pcs[0])
		r.AddAttrs(
			slog.String("url", req.URL.String()),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration),
			bodyAsAttrs(resp),
		)
		_ = l.Handler().Handle(req.Context(), r)
		return nil, err
	}

	r = slog.NewRecord(time.Now(), slog.LevelDebug, "http response", pcs[0])
	r.AddAttrs(
		slog.String("url", req.URL.String()),
		slog.Int("status", resp.StatusCode),
		slog.Duration("duration", duration),
		slog.Int64("content-length", resp.ContentLength),
		slog.Group("headers", headersAsAttrs(resp.Header)...),
	)

	if resp.StatusCode >= 400 {
		r.AddAttrs(bodyAsAttrs(resp))
	}

	_ = l.Handler().Handle(req.Context(), r)

	return resp, nil
}

func bodyAsAttrs(resp *http.Response) slog.Attr {
	if resp == nil || resp.Body == nil {
		return slog.String("body", "no body")
	}

	var buf bytes.Buffer
	_, err := io.Copy(&buf, resp.Body)
	if err != nil {
		return slog.String("body", "failed to read body")
	}

	resp.Body = io.NopCloser(&buf)

	if resp.Header.Get("Content-Type") == "application/json" {
		var data any
		if err := json.NewDecoder(&buf).Decode(&data); err != nil {
			return slog.String("body", "failed to decode body")
		}
		return slog.Any("body", data)
	}
	return slog.String("body", buf.String())
}

func headersAsAttrs(headers http.Header) []any {
	attrs := make([]any, 0, len(headers))
	for k, v := range headers {
		attrs = append(attrs, slog.String(k, strings.Join(v, ", ")))
	}
	return attrs
}
