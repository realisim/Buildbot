package buildbot

import (
	"fmt"
)

type qtInstaller struct {
	qt *qtConfig
	t  *target
}

func (i qtInstaller) createInstaller() (installerArtefact, error) {
	fmt.Printf("Creating qt installer for target %v...\n", i.t.Name)

	var err error
	return installerArtefact{}, err
}
