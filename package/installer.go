package buildbot

import (
	"fmt"
)

type installerArtefact struct {
	filePath, fileName string
}

type installer interface {
	createInstaller() (installerArtefact, error) //creates and return the path to the installer file
}

type noInstaller struct {
	t *target
}

func (i noInstaller) createInstaller() (installerArtefact, error) {
	fmt.Printf("Creating no installer for target %v...\n", i.t.Name)

	var err error
	return installerArtefact{}, err
}
