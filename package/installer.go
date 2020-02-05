package buildbot

import (
	"fmt"
)

type installer interface {
	createInstaller(outputPath string) error //creates and return the path to the installer file
}

type noInstaller struct {
	t *target
}

func (i noInstaller) createInstaller(outputPath string) error {
	fmt.Printf("Creating no installer for target %v...\n", i.t.Name)

	var err error
	return err
}
