package main

import (
	"flag"
	"fmt"
	"github.com/realisim/buildbot/package"
)

func main() {
	configFlagPtr := flag.String("config", "config.json", "Specifies the filepath to the config file.")
	targetFlagPtr := flag.String("target", "", "Specifies the target to build. Targets are defined by the config file. By default, all targets are built.")
	makeTemplateConfigFlagPtr := flag.Bool("makeTemplaceConfig", false, "Create and saves an empty templateConfig named templateConfig.json")

	flag.Parse()

	if *makeTemplateConfigFlagPtr {
		err := buildbot.MakeTemplateConfig()
		if err != nil {
			fmt.Printf("Could not create template config: %v\n", err)
		}
	}

	buildbot.Build(*configFlagPtr, *targetFlagPtr)
}
