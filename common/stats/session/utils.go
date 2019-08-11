package session

import (
	"fmt"
	"strings"
	"time"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func date(t time.Time) string {
	return t.Format("Mon Jan 2 15:04:05")
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func uptime() string {
	// Y M d h m s
	now := time.Now()
	year, month, day, hour, min, sec := diff(startTime, now)

	var Y, M, d, h, m, s string

	// Y M d
	if year != 0 {
		Y = fmt.Sprintf("%dY,", year)
	}

	if month != 0 {
		M = fmt.Sprintf("%dM,", month)
	}

	if day != 0 {
		d = fmt.Sprintf("%dd,", day)
	}

	// h m s
	if hour != 0 {
		h = fmt.Sprintf("%dh", hour)
	}

	if min != 0 {
		m = fmt.Sprintf("%dm", min)
	}

	if sec != 0 {
		s = fmt.Sprintf("%ds", sec)
	}

	return strings.Join([]string{Y, M, d, h, m, s}, "")
}

func diff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}
