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
	QtBinPath                   string
	QtLibPath                   string
	QtInstallerFrameworkBinPath string
}

type buildType string

const (
	debugBuild                buildType = "debugBuild"
	releaseBuild                        = "releaseBuild"
	releaseWithDebugInfoBuild           = "releaseWithDebugInfoBuild"
)

type installerType string

const (
	noInstallerType  = "noInstaller"
	zipInstallerType = "zipInstaller"
	qtInstallerType  = "qtInstaller"
)

type qtInstallerConfig struct {
	QtInstallerFolderPath                            string
	BuildArtefactToQtInstallerPackagesRelativeFolder map[string]string
}

type target struct {
	Name string

	// installer related
	RequiresQtDeploy   bool
	ArtefactFolderPath string
	ArtefactFileName   string
	InstallerType      installerType
	QtInstallerConfig  *qtInstallerConfig

	VersionFilePath string
	versionTuple    [4]int
}

type build struct {
	Name string

	BuildType         buildType
	CleanBeforeBuild  bool
	CmakeBuildOptions []string

	TargetNames []string
}

type config struct {
	DeployPath string
	Repo       repository
	Git        gitSourceControl
	Cmake      cmakeGenerator
	Qt         qtConfig
	Builds     []build
	Targets    map[string]target
}

// this function will create and save a template config file name templateConfig.json
//
func MakeTemplateConfig() error {
	var c config

	c.Cmake.Generator = "Visual Studio 15 2017 Win64"

	// make targets
	t0 := target{
		Name:               "DummyName",
		ArtefactFolderPath: "d:/somePath",
		InstallerType:      "zipInstaller",
		VersionFilePath:    "c:/pathTo/SomeVersion/File.h"}
	t1 := target{
		Name:               "DummyName2",
		ArtefactFolderPath: "d:/somePath2",
		InstallerType:      "zipInstaller",
		VersionFilePath:    "c:/pathTo/SomeVersion/File.h"}
	t2 := target{
		Name:               "DummyName3",
		ArtefactFolderPath: "d:/somePath3",
		InstallerType:      "qtInstaller",
		VersionFilePath:    ""} // no path means no version increment

	t2.QtInstallerConfig = &qtInstallerConfig{
		QtInstallerFolderPath: "e:/installer",
		BuildArtefactToQtInstallerPackagesRelativeFolder: map[string]string{
			"d:/path/to/my/build/artefact":  "packages/com.mycompany.myexec/data",
			"d:/path/to/my/build/artefact2": "packages/com.mycompany.myexec.Options/data"},
	}

	c.Targets = make(map[string]target)
	c.Targets[t0.Name] = t0
	c.Targets[t1.Name] = t1
	c.Targets[t2.Name] = t2

	//make bbuild
	b := build{
		Name:              "Dummy debugbuild",
		BuildType:         "debugBuild",
		CleanBeforeBuild:  false,
		CmakeBuildOptions: []string{"-DOption0=1", "-DOption1=1"}}

	b.TargetNames = []string{t0.Name, t1.Name}

	// build 2
	b2 := build{
		Name:              "Dummy release build",
		BuildType:         "releaseWithDebugInfoBuild",
		CleanBeforeBuild:  true,
		CmakeBuildOptions: []string{"-DOption0=1", "-DOption1=0"}}

	b2.TargetNames = []string{t0.Name, t1.Name, t2.Name}

	c.Builds = []build{b, b2}

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
