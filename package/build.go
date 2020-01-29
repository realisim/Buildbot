package buildbot

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
)

func gitUpdate(configPtr *config) error {
	// checkout master branch
	checkoutMaster := exec.Command(configPtr.Git.ExePath, "checkout", "master")
	checkoutMaster.Dir = configPtr.Repo.Path

	_, err := commandOutput(checkoutMaster)
	if err != nil {
		return fmt.Errorf("checkout failed: %v", err)
	}

	//--- pull
	pullMaster := exec.Command(configPtr.Git.ExePath, "pull")
	pullMaster.Dir = configPtr.Repo.Path

	_, err = commandOutput(pullMaster)
	if err != nil {
		fmt.Errorf("pull failed: %v", err)
	}

	return nil
}

func Build(iConfigFilePath, iTarget string) {
	fmt.Printf("Build call with config %s and taget %s\n", iConfigFilePath, iTarget)

	c, err := parseConfig(iConfigFilePath)
	if err != nil {
		fmt.Printf("Parse config failed: %v\n", err)
		return
	}

	// git update
	if c.Repo.UpdateBeforeBuild {
		if err := gitUpdate(&c); err != nil {
			fmt.Printf("gitUpdate failed: %v\n", err)
			return
		}
	}

	// make an array of seleted target
	var selectedTargets []target
	if iTarget == "" {
		selectedTargets = c.Targets[:]
	} else {
		// find index of selected target
		for i := range c.Targets {
			if c.Targets[i].Name == iTarget {
				selectedTargets = append(selectedTargets, c.Targets[i])
			}
		}

	}

	// for each target
	for _, t := range selectedTargets {
		fmt.Printf("Building target %v\n", t)

		// cmake generator
		if err := cmakeGenerate(&c, &t); err != nil {
			fmt.Printf("cmakeGenerate failed: %v\n", err)
			return
		}

		//increment version
		incrementVersion(t)

		// cmake build
		if err := cmakeBuild(&c, &t); err != nil {
			fmt.Printf("cmakeBuild failed: %v\n", err)
			return
		}

		// cmake run install target
		if err := cmakeInstall(&c, &t); err != nil {
			fmt.Printf("cmakeInstall failed: %v\n", err)
			return
		}

		// create installer

		// deploy
	}

	fmt.Printf("Build done.")
}

func incrementVersion(iTarget target) {
	if iTarget.VersionFilePath == "" {
		return
	}

	versionFile, err := ioutil.ReadFile(iTarget.VersionFilePath)
	if err != nil {
		fmt.Printf("if this happens, investigate if we should treat it as an error...\n")
		return
	}

	regexpPatern := regexp.MustCompile("VERSION_BUILDNUMBER ([0-9]+)")
	matches := regexpPatern.FindSubmatch(versionFile)
	if len(matches) == 2 {
		matches[1], _ = strconv.Atoi((strconv.ParseInt(matches[1], 10, 32) + 1))
	}

	fmt.Printf("%v\n", versionFile)
}
