package engine

import (
	"runtime"
	"testing"
)

func TestFDOffset(t *testing.T) {
	want := 0
	if runtime.GOOS == "darwin" || runtime.GOOS == "ios" {
		want = 4
	}
	if got := fdOffset(); got != want {
		t.Errorf("fdOffset() = %d, want %d on %s", got, want, runtime.GOOS)
	}
}
