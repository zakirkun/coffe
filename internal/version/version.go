package version

import "strings"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func Semver() string {
	if Version == "" || Version == "dev" {
		return ""
	}
	if strings.HasPrefix(Version, "v") {
		return Version
	}
	return "v" + Version
}
