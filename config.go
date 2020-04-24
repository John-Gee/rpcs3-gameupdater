/* rpcs3-gameupdater - maintains app wide configuration and persistence */

package main

import (
	"os"
	"runtime"
)

type Verbosity int

const (
	None Verbosity = iota
	Error
	Warning
	Info
	Debug
)

// Config is the app wide configuration structure
type Config struct {
	Rpcs3Path     string
	PkgDLPath     string
	ConfigYMLPath string
	DLTimeout     int
	DLRetries     int
	verbosity     Verbosity
	color         bool
}

var conf Config
var confPath string

func fetchConfig() Config {
	// fetch config here
	return conf
}

func persistConfig() {
	// store config here

	return
}

// this sets up the initial configuration
func initConfig() {
	conf = Config{
		Rpcs3Path:     ".",
		PkgDLPath:     ".",
		ConfigYMLPath: "",
		DLTimeout:     30,
		DLRetries:     3,
		color:         true,
		verbosity:     Debug,
	}
	confFile := "/rpcs3/config.yml"
	goos := runtime.GOOS
	switch goos {
	case "freebsd":
		fallthrough
	case "linux":
		if home := os.Getenv("XDG_CONFIG_HOME"); home != "" {
			conf.ConfigYMLPath = home + confFile
		} else if home := os.Getenv("HOME"); home != "" {
			conf.ConfigYMLPath = home + "/.config" + confFile
		} else {
			conf.ConfigYMLPath = "~/.config" + confFile
		}
	case "windows":
		conf.ConfigYMLPath = os.Getenv("RPCS3_CONFIG_DIR") + confFile
	}
	printInfo("config.yml should be at: " + conf.ConfigYMLPath)
	confPath = "."
}

// TODO: handle config changes gracefully, from UI and from command line args
