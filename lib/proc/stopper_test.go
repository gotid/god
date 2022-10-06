package proc

import "testing"

func TestNopStopper(t *testing.T) {
	// no panic
	nopStopper.Stop()
}
