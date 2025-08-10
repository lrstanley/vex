// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package formatter

import (
	"fmt"
	"time"
)

// Time formats a time.Time as a full string (date, time, timezone).
func Time(t time.Time) string {
	if t.IsZero() {
		return "n/a"
	}
	return t.Format(time.UnixDate)
}

// TimeRelative formats a time.Time as a relative string (e.g. "in 10 seconds",
// "10 seconds ago", "now", "n/a" for zero value).
func TimeRelative(t time.Time, postfix bool) string {
	if t.IsZero() {
		return "n/a"
	}

	d := time.Until(t)
	if d > 0 {
		if !postfix {
			return Duration(d, 0)
		}
		return Duration(d, 1)
	}
	d = time.Since(t)
	if !postfix {
		return Duration(d, 0)
	}
	return Duration(d, -1)
}

// Duration formats a duration, and optionally adds a relative prefix/suffix. If
// rel is 0, then the duration is formatted without a relative prefix/suffix. If
// rel is 1, it is considered in the future. If rel is -1, it is considered in
// the past.
func Duration(d time.Duration, rel int) string {
	if d == 0 {
		return "now"
	}

	var out string

	switch {
	case d > 3*365*24*time.Hour: // 3 years.
		out = fmt.Sprintf("%d years", int64(d.Round(time.Hour).Hours()/24/365))
	case d > 3*30*24*time.Hour: // 90 days.
		out = fmt.Sprintf("%d months", int64(d.Round(time.Hour).Hours()/24/30))
	case d > 3*7*24*time.Hour: // 3 weeks.
		out = fmt.Sprintf("%d weeks", int64(d.Round(time.Hour).Hours()/24/7))
	case d > 3*24*time.Hour: // 3 days.
		out = fmt.Sprintf("%d days", int64(d.Round(time.Hour).Hours()/24))
	case d > 3*time.Hour: // 3 hours.
		out = fmt.Sprintf("%d hours", int64(d.Round(time.Minute).Minutes()/60))
	case d > 3*time.Minute: // 3 minutes.
		out = fmt.Sprintf("%d minutes", int64(d.Round(time.Second).Seconds()/60))
	case d > time.Second: // 1 second.
		out = fmt.Sprintf("%d seconds", int64(d.Round(time.Second).Seconds()))
	default:
		return "now"
	}

	if rel > 0 {
		out = "in " + out
	} else if rel < 0 {
		out += " ago"
	}

	return out
}
