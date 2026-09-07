package platform

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>
*/
import "C"

const ShiftPollingSupported = true

// IsShiftPressed polls the OS-level modifier key state via CoreGraphics.
// Returns true if either Shift key is currently held down.
// This works without any special permissions on macOS.
func IsShiftPressed() bool {
	flags := C.CGEventSourceFlagsState(C.kCGEventSourceStateCombinedSessionState)
	return flags&C.kCGEventFlagMaskShift != 0
}
