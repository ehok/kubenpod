package version

import (
	"fmt"
	"runtime"
)

var (
	gitVersion = "none"
	gitCommit  = "none"
	buildDate  = "none"
)

var ver = Version{
	GoVersion:  runtime.Version(),
	GoOs:       runtime.GOOS,
	GoArch:     runtime.GOARCH,
	GitVersion: gitVersion,
	GitCommit:  gitCommit,
	BuildDate:  buildDate,
}

type Version struct {
	GoVersion  string
	GoOs       string
	GoArch     string
	GitVersion string
	GitCommit  string
	BuildDate  string
}

func Get() Version {
	return ver
}

func PrintVersion() {
	v := Get()
	fmt.Printf("Version: %s\nGit commit: %s\nBuild date: %s\n", v.GitVersion, v.GitCommit, v.BuildDate)
}
