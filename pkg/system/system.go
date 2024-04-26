package system

import "path/filepath"

const (
	// DefaultLocalDir represents where llmos persistent installation is located
	DefaultLocalDir = "/var/lib/llmos"
	// DefaultConfigDir represents where persistent configuration is located
	DefaultConfigDir = "/etc/llmos"
	// DefaultStateDir represents where cos ephemeral state is located
	DefaultStateDir = "/run/cos"
)

var (
	localDir  = DefaultLocalDir
	configDir = DefaultConfigDir
	stateDir  = DefaultStateDir

	// LocalConfigs represents the local llmos configuration directory
	LocalConfigs = ConfigPath("config.d")
)

func LocalPath(elem ...string) string {
	return filepath.Join(localDir, filepath.Join(elem...))
}

func ConfigPath(elem ...string) string {
	return filepath.Join(configDir, filepath.Join(elem...))
}

func StatePath(elem ...string) string {
	return filepath.Join(stateDir, filepath.Join(elem...))
}
