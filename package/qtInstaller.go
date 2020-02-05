package buildbot

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"path/filepath"
)

type qtInstaller struct {
	c *config
	t  *target
}

func (i qtInstaller) createInstaller(outputPath string) error {
	fmt.Printf("Creating qt installer for target %v...\n", i.t.Name)

	tempFolderPath := "c:\\temp\\installer"

	var qic *qtInstallerConfig
	qic = i.t.QtInstallerConfig
	if qic != nil {
		// copy installer code to c:/temp/installer
		// this way we do not mess up the build tree
		// the folder will be removed at the end
		//
		args := []string{ filepath.Join(i.c.Repo.Path, qic.QtInstallerRelativeFolderPath), tempFolderPath, "/E"}
		command := exec.Command("robocopy", args...)
		commandOutput(command)

		// copy build artefact to installer/package/data
		for k, v := range qic.BuildArtefactToQtInstallerPackagesRelativeFolder {
			args := []string{
				filepath.Join(i.c.Repo.Path, k),
				fmt.Sprintf("%v/%v", tempFolderPath, v),
				"/E"}
			command := exec.Command("robocopy", args...)
			commandOutput(command)

			// update version in config
			configXmlFilePath := tempFolderPath + "/config/config.xml"
			i.updateVersion(configXmlFilePath)

			// create installer
			finalInstallerFileName := fmt.Sprintf("%v/%v_%v.exe", outputPath, i.t.ArtefactFileName, i.t.versionTuple)
			args = []string{"-c", configXmlFilePath, "-p", tempFolderPath + "/packages", finalInstallerFileName}
			command = exec.Command(i.c.Qt.QtInstallerFrameworkBinPath+"/binarycreator.exe",
				args...)
			commandOutput(command)

			//delete temp data
			if err := os.RemoveAll(tempFolderPath); err != nil {
				return fmt.Errorf("Could not remove temp folder %v\n", tempFolderPath)
			}
		}
	}

	var err error
	return err
}

func (i qtInstaller) updateVersion(f string) error {
	var qic *qtInstallerConfig
	qic = i.t.QtInstallerConfig

	if qic == nil {
		return nil
	}

	content, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("could not read config.xml file...:%v\n", err)
	}

	regexPatern := regexp.MustCompile(`(<Version>)([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)(</Version>)`)
	matches := regexPatern.FindSubmatch(content)
	if len(matches) == 4 {
		// replace in file
		newVersionString := fmt.Sprintf("<Version>%d.%d.%d.%d</Version>",
			i.t.versionTuple[0],
			i.t.versionTuple[1],
			i.t.versionTuple[2],
			i.t.versionTuple[3])
		newContent := strings.Replace(string(content),
			string(matches[0]),
			newVersionString, 1)

		ioutil.WriteFile(f, []byte(newContent), 0)

		fmt.Printf("Updated to version %v in file %v\n", newVersionString, f)
	}
	return nil
}
