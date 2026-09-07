package platform

import (
	"encoding/base64"
	"fmt"
	"io"
)

// CopyToClipboard writes an OSC 52 clipboard sequence to the TUI output.
func CopyToClipboard(output io.Writer, value string) error {
	if output == nil {
		return fmt.Errorf("clipboard output is not configured")
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(value))
	_, err := fmt.Fprintf(output, "\x1b]52;c;%s\x07", encoded)
	return err
}
