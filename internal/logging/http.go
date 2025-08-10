// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package logging

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strconv"
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

	trace, _ := strconv.ParseBool(os.Getenv("HTTP_TRACE"))

	var pcs [1]uintptr
	_ = runtime.Callers(6, pcs[:]) // Skip this, and all of the net/http/client functions.

	r := slog.NewRecord(time.Now(), slog.LevelDebug, "http request", pcs[0])

	r.AddAttrs(
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.String("user-agent", req.UserAgent()),
		slog.Int64("content-length", req.ContentLength),
	)

	if trace {
		b, err := httputil.DumpRequest(req, true)
		if err == nil {
			r.AddAttrs(slog.String("request", string(b)))
		}
	}

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
		)

		if resp != nil && trace {
			var b []byte
			b, err = httputil.DumpResponse(resp, true)
			if err == nil {
				r.AddAttrs(slog.String("response", string(b)))
			}
		}

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

	if trace {
		var b []byte
		b, err = httputil.DumpResponse(resp, true)
		if err == nil {
			r.AddAttrs(slog.String("response", string(b)))
		}
	}

	_ = l.Handler().Handle(req.Context(), r)

	return resp, nil
}

func headersAsAttrs(headers http.Header) []any {
	attrs := make([]any, 0, len(headers))
	for k, v := range headers {
		attrs = append(attrs, slog.String(k, strings.Join(v, ", ")))
	}
	return attrs
}
