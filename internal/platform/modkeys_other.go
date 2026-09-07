//go:build !darwin && !windows

package platform

const ShiftPollingSupported = false

// IsShiftPressed is a no-op on Linux and other platforms.
// Shift detection there relies on the Kitty keyboard protocol instead.
func IsShiftPressed() bool {
	return false
}
