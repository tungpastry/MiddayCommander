package completion

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCurrentWordAndCommonPrefix(t *testing.T) {
	start, end, word := CurrentWord("cp alpha/beta tail", len("cp alpha/be"))
	if start != 3 || end != len("cp alpha/beta") || word != "alpha/beta" {
		t.Fatalf("CurrentWord = (%d, %d, %q)", start, end, word)
	}
	if got := CommonPrefix([]string{"alpha", "alpine", "alps"}); got != "alp" {
		t.Fatalf("CommonPrefix = %q, want alp", got)
	}
}

func TestCompletePathCandidates(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "draft.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := CompletePathCandidates("d", root, false); !reflect.DeepEqual(got, []string{"docs/", "draft.txt"}) {
		t.Fatalf("path candidates = %#v", got)
	}
	if got := CompletePathCandidates("d", root, true); !reflect.DeepEqual(got, []string{"docs/"}) {
		t.Fatalf("directory candidates = %#v", got)
	}
}

func TestCompleteExecCandidatesUsesPATHAndExecutableBit(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "mdc-tool"), nil, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "mdc-note"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", root)
	execCacheMu.Lock()
	execCache = nil
	execPathEnv = ""
	execCacheMu.Unlock()
	if got := CompleteExecCandidates("mdc-", root); !reflect.DeepEqual(got, []string{"mdc-tool"}) {
		t.Fatalf("executable candidates = %#v", got)
	}
}
