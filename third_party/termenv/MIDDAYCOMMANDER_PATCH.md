# MiddayCommander Termenv Patch

This directory is a source pin of `github.com/muesli/termenv` v0.16.0.
The upstream source and MIT license are preserved.

MiddayCommander changes only `OSCTimeout` in `termenv_unix.go`, reducing it
from 5 seconds to 100 milliseconds. Bubble Tea v1 queries the terminal's
background color during package initialization. Apple Terminal can ignore that
OSC query, which otherwise delays entry into `main` by the full timeout.

Remove the local replacement after migrating to Bubble Tea v2, where terminal
color requests are handled inside the event loop.
