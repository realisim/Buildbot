package main

import (
	"flag"
	"fmt"
	"github.com/realisim/buildbot/package"
)

func main() {
	configFlagPtr := flag.String("config", "config.json", "Specifies the filepath to the config file.")
	buildFlagPtr := flag.String("build", "", "Specifies the build. Builds are defined in the config file.")
	makeTemplateConfigFlagPtr := flag.Bool("makeTemplateConfig", false, "Create and saves an empty templateConfig named templateConfig.json")

	flag.Parse()

	if *makeTemplateConfigFlagPtr {
		err := buildbot.MakeTemplateConfig()
		if err != nil {
			fmt.Printf("Could not create template config: %v\n", err)
		}
	}

	if err := buildbot.Build(*configFlagPtr, *buildFlagPtr); err != nil {
		fmt.Printf("build failed: %v\n", err)
	}
}
