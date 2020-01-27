package buildbot

import (
	"fmt"
	"os"
	"os/exec"
)

func commandOutput(cmd *exec.Cmd) (string, error) {
	fmt.Printf("Command: %v\n", cmd.Args)

	var out []byte
	var err error
	if out, err = cmd.Output(); err != nil {
		return string(out), fmt.Errorf("command failed: %v", err)
	}

	fmt.Printf(string(out))
	return string(out), nil
}

func cmakeBuild(configPtr *config, t *target) error {
	buildType := toCmakeBuildType(t.BuildType)

	args := []string{"--build", configPtr.Repo.BuildPath,
		"--config", buildType}
	if t.CleanBeforeBuild {
		args = append(args, "--clean-first")
	}

	buildCommand := exec.Command(configPtr.Cmake.ExePath,
		args...)

	buildCommand.Dir = configPtr.Repo.BuildPath

	var out string
	var err error
	if out, err = commandOutput(buildCommand); err != nil {
		return fmt.Errorf("cmakeBuild failed: %v\n, %v", err, out)
	}

	return nil
}

func cmakeGenerate(configPtr *config, t *target) error {

	// start by removing the CMakeCache.txt as it contains previous options
	// set by someone or another build..
	cacheFilePath := configPtr.Repo.BuildPath + "/CMakeCache.txt"
	if _, err := os.Stat(cacheFilePath); err == nil {
		// file exists, remove it
		os.Remove(cacheFilePath)
	}

	// create build path if not exists
	if _, err := os.Stat(configPtr.Repo.BuildPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configPtr.Repo.BuildPath, 0777); err != nil {
			return fmt.Errorf("could not create build folder %v, %v", configPtr.Repo.BuildPath, err)
		}
	}

	// run the cmake command in the build path
	// la commande a passer pour cmake est dans le genre
	//cmake ../ -G 'Visual Studio 14 2015 Win64'
	//-DCMAKE_PREFIX_PATH=$QT_PREFIX_PATH/lib/cmake/Qt5OpenGL;$QT_PREFIX_PATH/lib/cmake/Qt5Widgets;$QT_PREFIX_PATH/lib/cmake/Qt5PrintSupport;$QT_PREFIX_PATH/lib/cmake/Qt5Network

	qtConfig := fmt.Sprintf("-DCMAKE_PREFIX_PATH=%v/cmake/Qt5OpenGL;%v/cmake/Qt5Widgets;%v/cmake/Qt5Core;%v/cmake/Qt5Gui",
		configPtr.Qt.QtLibPath, configPtr.Qt.QtLibPath, configPtr.Qt.QtLibPath, configPtr.Qt.QtLibPath)

	args := []string{"../",
		"-G", configPtr.Cmake.Generator}
	args = append(args, t.CmakeBuildOptions...)
	args = append(args, qtConfig)

	generate := exec.Command(configPtr.Cmake.ExePath, args...)
	generate.Dir = configPtr.Repo.BuildPath

	if _, err := commandOutput(generate); err != nil {
		return fmt.Errorf("cmakeGenerate failed: %v", err)
	}

	return nil
}

func cmakeInstall(configPtr *config, t *target) error {
	buildType := toCmakeBuildType(t.BuildType)
	buildCommand := exec.Command(configPtr.Cmake.ExePath,
		"--build", configPtr.Repo.BuildPath,
		"--config", buildType,
		"--target", "INSTALL")

	buildCommand.Dir = configPtr.Repo.BuildPath

	if _, err := commandOutput(buildCommand); err != nil {
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
