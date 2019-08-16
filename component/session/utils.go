package session

import (
	"fmt"
	"strings"
	"time"

	C "github.com/shirou/gopsutil/cpu"
	H "github.com/shirou/gopsutil/host"
	M "github.com/shirou/gopsutil/mem"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func cpu() string {
	c, err := C.Percent(0, false)
	if err != nil || len(c) != 1 {
		return "N/A"
	}
	return fmt.Sprintf("%.1f%%", c[0])
}

func platform() string {
	h, err := H.Info()
	if err != nil {
		return "N/A"
	}
	return fmt.Sprintf("%s-%s", h.Platform, h.KernelVersion)
}

func mem() string {
	m, err := M.VirtualMemory()
	if err != nil {
		return "N/A"
	}
	return fmt.Sprintf("%.1f%%", m.UsedPercent)
}

func date(t time.Time) string {
	return t.Format("Mon Jan 2 15:04:05")
}

func duration(start, end time.Time) string {
	var t time.Duration
	if end.IsZero() {
		t = time.Now().Sub(start)
	} else {
		t = end.Sub(start)
	}

	switch {
	case t < 1000*time.Millisecond:
		t = t.Round(time.Millisecond)
	default:
		t = t.Round(time.Second)
	}
	return t.String()
}

func uptime() string {
	// Time difference function
	diff := func(a, b time.Time) (year, month, day, hour, min, sec int) {
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

func byteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
