package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time using ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info contains version information
type Info struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Platform  string
}

// GetVersionInfo returns version information
func GetVersionInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// String returns a formatted version string
func (v Info) String() string {
	if v.Version == "dev" {
		return fmt.Sprintf("%s (dev, commit %s)", v.Version, v.Commit)
	}
	return fmt.Sprintf("%s (commit %s, built %s)", v.Version, v.Commit, v.BuildDate)
}

// FullString returns a detailed version string
func (v Info) FullString() string {
	return fmt.Sprintf(`legionbatctl %s
Commit:    %s
Built:     %s
Go:        %s
Platform:  %s`,
		v.Version, v.Commit, v.BuildDate, v.GoVersion, v.Platform)
}
