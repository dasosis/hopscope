// Package version reports build metadata for the hopscope binary.
//
// Release builds inject Version, Commit and Date with -ldflags "-X ...". A
// plain `go build` or `go install` leaves them empty, so Get falls back to the
// module and VCS information the Go toolchain embeds in every binary.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Set at link time, e.g.
//
//	go build -ldflags "-X github.com/dasosis/hopscope/internal/version.Version=v0.1.0"
var (
	Version = ""
	Commit  = ""
	Date    = ""
)

// Info describes the running binary.
type Info struct {
	Version   string
	Commit    string
	Date      string
	GoVersion string
	Platform  string
}

// Get returns build information, preferring link-time values and falling back
// to the build info embedded by the Go toolchain.
func Get() Info {
	info := Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		if info.Version == "" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			info.Version = bi.Main.Version
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if info.Commit == "" {
					info.Commit = s.Value
				}
			case "vcs.time":
				if info.Date == "" {
					info.Date = s.Value
				}
			}
		}
	}

	if info.Version == "" {
		info.Version = "dev"
	}
	if len(info.Commit) > 12 {
		info.Commit = info.Commit[:12]
	}
	return info
}

// String renders Info as a single human-readable line.
func (i Info) String() string {
	s := "hopscope " + i.Version
	if i.Commit != "" {
		s += " (" + i.Commit
		if i.Date != "" {
			s += ", " + i.Date
		}
		s += ")"
	}
	return fmt.Sprintf("%s %s %s", s, i.GoVersion, i.Platform)
}
