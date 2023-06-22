package main

import (
	"runtime/debug"

	"github.com/vmkteam/mfd-generator/cmd"
	"github.com/vmkteam/mfd-generator/mfd"
)

func main() {
	setAppVersion()
	cmd.Execute()
}

func setAppVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	mfd.Version = info.Main.Version
}
