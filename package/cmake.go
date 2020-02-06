package buildbot

import (
	"fmt"
	"os"
	"os/exec"
)

type cmake struct {
	configPtr *config
	buildPtr  *build
}

func commandOutput(cmd *exec.Cmd) error {
	fmt.Printf("Command: %v\n", cmd.Args)

	var out []byte
	var err error
	if out, err = cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("command failed: %v\n, %v\n", err, string(out))
	}

	fmt.Printf(string(out))
	return nil
}

func (c cmake) Build() error {
	buildType := toCmakeBuildType(c.buildPtr.BuildType)

	args := []string{"--build", 
		c.configPtr.Repo.BuildPath,
		"--config", buildType}
	if c.buildPtr.CleanBeforeBuild {
		args = append(args, "--clean-first")
	}

	buildCommand := exec.Command(c.configPtr.Cmake.ExePath,
		args...)

	buildCommand.Dir = c.configPtr.Repo.BuildPath

	if err := commandOutput(buildCommand); err != nil {
		return fmt.Errorf("cmakeBuild failed: %v\n", err)
	}

	return nil
}

func (c cmake) Generate() error {

	// start by removing the CMakeCache.txt as it contains previous options
	// set by someone or another build..
	cacheFilePath := c.configPtr.Repo.BuildPath + "/CMakeCache.txt"
	if _, err := os.Stat(cacheFilePath); err == nil {
		// file exists, remove it
		os.Remove(cacheFilePath)
	}

	// create build path if not exists
	if _, err := os.Stat(c.configPtr.Repo.BuildPath); os.IsNotExist(err) {
		if err := os.MkdirAll(c.configPtr.Repo.BuildPath, 0777); err != nil {
			return fmt.Errorf("could not create build folder %v, %v", c.configPtr.Repo.BuildPath, err)
		}
	}

	// run the cmake command in the build path
	// la commande a passer pour cmake est dans le genre
	//cmake ../ -G 'Visual Studio 14 2015 Win64'
	//-DCMAKE_PREFIX_PATH=$QT_PREFIX_PATH/lib/cmake/Qt5OpenGL;$QT_PREFIX_PATH/lib/cmake/Qt5Widgets;$QT_PREFIX_PATH/lib/cmake/Qt5PrintSupport;$QT_PREFIX_PATH/lib/cmake/Qt5Network

	qtConfig := fmt.Sprintf("-DCMAKE_PREFIX_PATH=%v/cmake/Qt5OpenGL;%v/cmake/Qt5Widgets;%v/cmake/Qt5Core;%v/cmake/Qt5Gui",
		c.configPtr.Qt.QtLibPath, c.configPtr.Qt.QtLibPath, c.configPtr.Qt.QtLibPath, c.configPtr.Qt.QtLibPath)

	args := []string{"../",
		"-G", c.configPtr.Cmake.Generator}
	args = append(args, c.buildPtr.CmakeBuildOptions...)
	args = append(args, qtConfig)

	generate := exec.Command(c.configPtr.Cmake.ExePath, args...)
	generate.Dir = c.configPtr.Repo.BuildPath

	if err := commandOutput(generate); err != nil {
		return fmt.Errorf("cmakeGenerate failed: %v", err)
	}

	return nil
}

func (c cmake) Install() error {
	buildType := toCmakeBuildType(c.buildPtr.BuildType)
	buildCommand := exec.Command(c.configPtr.Cmake.ExePath,
		"--build", c.configPtr.Repo.BuildPath,
		"--config", buildType,
		"--target", "INSTALL")

	buildCommand.Dir = c.configPtr.Repo.BuildPath

	if err := commandOutput(buildCommand); err != nil {
		return fmt.Errorf("cmakeInstall failed: %v", err)
	}

	return nil
}

func toCmakeBuildType(bt buildType) string {
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
