package platform

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestCopyToClipboardWritesOSC52(t *testing.T) {
	var output bytes.Buffer
	if err := CopyToClipboard(&output, "a path/file.txt"); err != nil {
		t.Fatal(err)
	}
	want := "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte("a path/file.txt")) + "\a"
	if output.String() != want {
		t.Fatalf("OSC52 = %q, want %q", output.String(), want)
	}
}
