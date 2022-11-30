package main

import (
	"github.com/linuxsuren/gogit/cmd"
	"os"
)

func main() {
	if err := cmd.NewBuildCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
