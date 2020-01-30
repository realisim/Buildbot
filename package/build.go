package buildbot

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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
		incrementVersion(&t)

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
		var i installer
		switch t.InstallerType {
		case noInstallerType:
			i = noInstaller{&t}
		case zipInstallerType:
			i = zipInstaller{&t}
		case qtInstallerType:
			i = qtInstaller{&c.Qt, &t}
		default:
		}

		var ia installerArtefact
		var err error
		if ia, err = i.createInstaller(); err != nil {
			fmt.Printf("creating installer failed: %v\n", err)
		}

		// deploy
		if err := os.Rename(ia.filePath, fmt.Sprintf("%v/%v", c.DeployPath, ia.fileName)); err != nil {
			fmt.Printf("could not deploy: %v\n", err)
		}

	}

	fmt.Printf("Build done.")
}

func incrementVersion(t *target) {
	if t.VersionFilePath == "" {
		return
	}

	content, err := ioutil.ReadFile(t.VersionFilePath)
	if err != nil {
		fmt.Printf("if this happens, investigate if we should treat it as an error...\n")
		return
	}

	// get full version tuble

	regexpPatern := regexp.MustCompile("(VERSION_MAJOR )([0-9]+).*\n.*(VERSION_MINOR )([0-9]+).*\n.*(VERSION_REVISION )([0-9]+).*\n.*(VERSION_BUILDNUMBER )([0-9]+)")
	matches := regexpPatern.FindSubmatch(content)
	if len(matches) == 9 {
		major, _ := strconv.ParseInt(string(matches[2]), 10, 64)
		minor, _ := strconv.ParseInt(string(matches[4]), 10, 64)
		revision, _ := strconv.ParseInt(string(matches[6]), 10, 64)
		buildNumber, _ := strconv.ParseInt(string(matches[8]), 10, 64)
		buildNumber += 1

		// assign version to target
		t.versionTuple = [4]int{int(major), int(minor), int(revision), int(buildNumber)}

		// replace in file
		newContent := strings.Replace(string(content),
			string(matches[7])+string(matches[8]),
			fmt.Sprintf("VERSION_BUILDNUMBER %d", buildNumber), 1)

		ioutil.WriteFile(t.VersionFilePath, []byte(newContent), 0)

		fmt.Printf("Build number increased from: %v.%v.%v.%v to %v.%v.%v.%v\n",
			major, minor, revision, buildNumber-1,
			major, minor, revision, buildNumber)
		fmt.Printf("%v\n", newContent)
	}
}
