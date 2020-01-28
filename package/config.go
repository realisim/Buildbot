package buildbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type repository struct {
	Path              string
	BuildPath         string
	UpdateBeforeBuild bool
}

type gitSourceControl struct {
	ExePath string
}

type cmakeGenerator struct {
	ExePath   string
	Generator string
	//BuildOptions []string
}

type qtConfig struct {
	QtBinPath                string
	QtLibPath                string
	QtInstallerFrameworkPath string
}

type buildType string

const (
	debugBuild                buildType = "debugBuild"
	releaseBuild                        = "releaseBuild"
	releaseWithDebugInfoBuild           = "releaseWithDebugInfoBuild"
)

type installerType string

const (
	noInstaller  = "noInstaller"
	zipInstaller = "zipInstaller"
	qtInstaller  = "qtInstaller"
)

type target struct {
	Name               string
	BuildType          buildType
	CleanBeforeBuild   bool
	CmakeBuildOptions  []string
	ArtefactFolderPath string
	InstallerType      installerType
}

type config struct {
	DeployPath string
	Repo       repository
	Git        gitSourceControl
	Cmake      cmakeGenerator
	Qt         qtConfig
	Targets    []target
}

// this function will create and save a template config file name templateConfig.json
//
func MakeTemplateConfig() error {
	var c config

	c.Cmake.Generator = "Visual Studio 15 2017 Win64"

	t0 := target{
		Name:               "DummyName",
		BuildType:          "debugBuild",
		CleanBeforeBuild:   false,
		CmakeBuildOptions:  []string{"-DOption0=1", "-DOption1=1"},
		ArtefactFolderPath: "d:/somePath",
		InstallerType:      "zipInstaller"}

	t1 := target{
		Name:               "DummyName2",
		BuildType:          "releaseWithDebugInfoBuild",
		CleanBeforeBuild:   true,
		CmakeBuildOptions:  []string{"-DOption0=1", "-DOption1=0"},
		ArtefactFolderPath: "d:/somePath",
		InstallerType:      "qtInstaller"}
	c.Targets = append(c.Targets, t0)
	c.Targets = append(c.Targets, t1)

	var prettyJson bytes.Buffer
	jsonBytes, _ := json.Marshal(&c)
	if error := json.Indent(&prettyJson, jsonBytes, "", "\t"); error != nil {
		return fmt.Errorf("MakeTemplateConfig failed")
	}

	ioutil.WriteFile("templateConfig.json", prettyJson.Bytes(), 0777)

	// print the json on screen for debug purposes
	fmt.Printf("---\n%v\n---\n", string(prettyJson.Bytes()))

	return nil
}

func parseConfig(iConfigFilePath string) (conf config, err error) {
	var c config

	if _, err := os.Stat(iConfigFilePath); err != nil {
		return c, fmt.Errorf("config file %v not found", iConfigFilePath)
	}

	jsonData, err := ioutil.ReadFile(iConfigFilePath)
	if err != nil {
		return c, fmt.Errorf("ioutil.ReadFile: %v", err)
	}

	if err := json.Unmarshal(jsonData, &c); err != nil {
		return c, fmt.Errorf("json.Unmarshal: %v", err)
	}

	return c, nil
}
