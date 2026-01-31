package version

// Version is set via ldflags at build time
var Version = "dev"

// GetVersion returns the current version of the agent
func GetVersion() string {
	return Version
}
