package cli

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/dasosis/hopscope/internal/version"
)

func TestVersionCommand(t *testing.T) {
	// Simulate a release build that had its metadata injected via -ldflags.
	setVar(t, &version.Version, "v1.2.3")
	setVar(t, &version.Commit, "0123456789abcdef")
	setVar(t, &version.Date, "2026-10-22T00:00:00Z")

	var out bytes.Buffer
	cmd := NewRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := "hopscope v1.2.3 (0123456789ab, 2026-10-22T00:00:00Z) " +
		runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH
	if got != want {
		t.Errorf("unexpected output\n got: %q\nwant: %q", got, want)
	}
}

// setVar overrides a package-level string for the duration of one test.
func setVar(t *testing.T, p *string, v string) {
	t.Helper()
	old := *p
	*p = v
	t.Cleanup(func() { *p = old })
}
