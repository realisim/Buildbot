package buildbot

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func gitUpdate(configPtr *config) error {
	// checkout master branch
	checkoutMaster := exec.Command(configPtr.Git.ExePath, "checkout", "master")
	checkoutMaster.Dir = configPtr.Repo.Path

	err := commandOutput(checkoutMaster)
	if err != nil {
		return fmt.Errorf("checkout failed: %v", err)
	}

	//--- pull
	pullMaster := exec.Command(configPtr.Git.ExePath, "pull")
	pullMaster.Dir = configPtr.Repo.Path

	err = commandOutput(pullMaster)
	if err != nil {
		fmt.Errorf("pull failed: %v", err)
	}

	return nil
}

func Build(iConfigFilePath, iBuild string) {
	fmt.Printf("Build call with config %s and build %s\n", iConfigFilePath, iBuild)

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
	var b build
	// find index of selected target
	for i := range c.Builds {
		if c.Builds[i].Name == iBuild {
			b = c.Builds[i]
		}
	}

	fmt.Printf("Building target %v\n", b)

	// cmake generator
	cmake := cmake{&c, &b}
	if err := cmake.Generate(); err != nil {
		fmt.Printf("cmakeGenerate failed: %v\n", err)
		return
	}

	//increment version
	incrementVersion(&c, &b)

	// cmake build
	if err := cmake.Build(); err != nil {
		fmt.Printf("cmakeBuild failed: %v\n", err)
		return
	}

	// cmake run install target
	if err := cmake.Install(); err != nil {
		fmt.Printf("cmakeInstall failed: %v\n", err)
		return
	}

	// deploy qt if necessary
	deployQt(&c, &b)

	// create installer for each target of build
	if err := createInstallers(&c, &b); err != nil {
		fmt.Printf("createInstallers failed: %v\n", err)
		return
	}

	fmt.Printf("Build done.")
}

func createInstallers(c *config, b *build) error {

	for _, tn := range b.TargetNames {
		t := c.Targets[tn]

		var i installer
		switch t.InstallerType {
		case noInstallerType:
			i = noInstaller{&t}
		case zipInstallerType:
			i = zipInstaller{c, &t}
		case qtInstallerType:
			i = qtInstaller{c, &t}
		default:
		}

		// log the error
		if err := i.createInstaller(c.DeployPath); err != nil {
			return fmt.Errorf("creating installer on target %v failed: %v\n", t.Name, err)
		}
	}

	return nil
}

func deployQt(c *config, b *build) error {
	for _, tn := range b.TargetNames {
		t := c.Targets[tn]

		if t.RequiresQtDeploy {

			args := []string{
				fmt.Sprintf("%v/%v/%v",
					t.artefactFolderPath(c.Repo.Path),
					toBuildFolderName(b.BuildType),
					t.ArtefactFileName),
				"--release", "--force"}

			command := exec.Command(c.Qt.QtBinPath+"/windeployqt.exe", args...)
			if err := commandOutput(command); err != nil {
				return fmt.Errorf("windeployqt.exe failed: %v", err)
			}
		}
	}

	return nil
}

// increment version file for each taget of build
func incrementVersion(c *config, b *build) {
	for _, tn := range b.TargetNames {
		t := c.Targets[tn]

		if t.RepoRelativeVersionFilePath == "" {
			return
		}

		content, err := ioutil.ReadFile(t.versionFilePath(c.Repo.Path))
		if err != nil {
			fmt.Printf("could not increment version: %v, if this happens, investigate if we should treat it as an error...\n", err)
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
			// reassign target to save versionTuple
			c.Targets[tn] = t

			// replace in file
			newContent := strings.Replace(string(content),
				string(matches[7])+string(matches[8]),
				fmt.Sprintf("VERSION_BUILDNUMBER %d", buildNumber), 1)

			ioutil.WriteFile(t.versionFilePath(c.Repo.Path), []byte(newContent), 0)

			fmt.Printf("Build number increased from: %v.%v.%v.%v to %v.%v.%v.%v\n",
				major, minor, revision, buildNumber-1,
				major, minor, revision, buildNumber)
			fmt.Printf("%v\n", newContent)
		}
	}
}

func toBuildFolderName(bt buildType) string {
	r := "Debug"
	switch bt {
	case debugBuild:
	case releaseBuild:
		r = "Release"
	case releaseWithDebugInfoBuild:
		r = "RelWithDebInfo"
	default:
	}
	return r
}
