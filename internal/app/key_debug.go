package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// KeyDebugLogger records key input diagnostics as JSON lines.
type KeyDebugLogger struct {
	mu     sync.Mutex
	writer io.Writer
	closer io.Closer
}

type keyDebugEvent map[string]any

// NewKeyDebugLogger opens path for append and returns a key diagnostic logger.
func NewKeyDebugLogger(path string) (*KeyDebugLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create key debug log dir: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open key debug log: %w", err)
	}
	return &KeyDebugLogger{writer: file, closer: file}, nil
}

func newKeyDebugLoggerForWriter(writer io.Writer) *KeyDebugLogger {
	return &KeyDebugLogger{writer: writer}
}

func (l *KeyDebugLogger) Close() error {
	if l == nil || l.closer == nil {
		return nil
	}
	return l.closer.Close()
}

func (l *KeyDebugLogger) log(event keyDebugEvent) {
	if l == nil || l.writer == nil {
		return
	}
	event["ts"] = time.Now().Format(time.RFC3339Nano)

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.writer.Write(append(data, '\n'))
}

func (l *KeyDebugLogger) LogShift(kind string, shiftHeld bool, lastShiftSeen time.Time) {
	l.log(keyDebugEvent{
		"event":           kind,
		"msg_type":        "app." + kind,
		"shift_held":      shiftHeld,
		"last_shift_seen": formatDebugTime(lastShiftSeen),
	})
}

func (l *KeyDebugLogger) LogMouse(msg tea.MouseMsg, shiftHeld bool, lastShiftSeen time.Time) {
	l.log(keyDebugEvent{
		"event":           "mouse",
		"msg_type":        fmt.Sprintf("%T", msg),
		"action":          fmt.Sprint(msg.Action),
		"button":          fmt.Sprint(msg.Button),
		"x":               msg.X,
		"y":               msg.Y,
		"mouse_shift":     msg.Shift,
		"shift_held":      shiftHeld,
		"last_shift_seen": formatDebugTime(lastShiftSeen),
	})
}

func (l *KeyDebugLogger) LogKey(msg, effective tea.KeyMsg, before keyDebugState, after keyDebugState, matchedAction string, dispatched bool) {
	l.log(keyDebugEvent{
		"event":              "key",
		"msg_type":           fmt.Sprintf("%T", msg),
		"received":           debugKey(msg),
		"effective":          debugKey(effective),
		"shift_held_before":  before.shiftHeld,
		"shift_held_after":   after.shiftHeld,
		"last_shift_before":  formatDebugTime(before.lastShiftSeen),
		"last_shift_after":   formatDebugTime(after.lastShiftSeen),
		"shift_recent":       after.shiftRecent,
		"matched_action":     matchedAction,
		"dispatched_global":  dispatched,
		"effective_changed":  msg.String() != effective.String() || tea.Key(msg).Type != tea.Key(effective).Type,
		"shift_grace_window": shiftFKeyGraceWindow.String(),
	})
}

func (l *KeyDebugLogger) LogKittyFilter(raw []byte, params string, converted tea.Msg) {
	event := keyDebugEvent{
		"event":      "kitty_filter",
		"msg_type":   "unknownCSISequenceMsg",
		"raw_hex":    hex.EncodeToString(raw),
		"raw_quoted": fmt.Sprintf("%q", string(raw)),
		"params":     params,
	}
	if converted != nil {
		event["converted_type"] = fmt.Sprintf("%T", converted)
		if keyMsg, ok := converted.(tea.KeyMsg); ok {
			event["converted_key"] = debugKey(keyMsg)
		}
	} else {
		event["converted_type"] = nil
	}
	l.log(event)
}

type keyDebugState struct {
	shiftHeld     bool
	lastShiftSeen time.Time
	shiftRecent   bool
}

func debugKey(msg tea.KeyMsg) map[string]any {
	keyMsg := tea.Key(msg)
	return map[string]any{
		"string":   msg.String(),
		"type":     fmt.Sprint(keyMsg.Type),
		"type_int": int(keyMsg.Type),
		"runes":    string(keyMsg.Runes),
		"alt":      keyMsg.Alt,
		"paste":    keyMsg.Paste,
	}
}

func formatDebugTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}
