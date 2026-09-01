package termenv

import (
	"testing"
	"time"
)

func TestMiddayCommanderOSCTimeout(t *testing.T) {
	const want = 100 * time.Millisecond
	if OSCTimeout != want {
		t.Fatalf("OSCTimeout = %s, want %s", OSCTimeout, want)
	}
}
