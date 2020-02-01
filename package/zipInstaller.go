package buildbot

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type zipInstaller struct {
	t *target
}

func (i zipInstaller) createInstaller(outputPath string) error {
	fmt.Printf("Creating zip installer for target %v...\n", i.t.Name)
	var err error

	var compressedBuffer = new(bytes.Buffer)

	w := zip.NewWriter(compressedBuffer)

	// compress everything in artefact path
	err = filepath.Walk(i.t.ArtefactFolderPath,
		func(path string, info os.FileInfo, err error) error {

			// do not archive folder...
			if info.IsDir() {
				return nil
			}

			relativePath := strings.TrimPrefix(path, filepath.Clean(i.t.ArtefactFolderPath)+"\\")
			fmt.Printf("Compressing %v | %v\n", path, relativePath)
			f, _ := w.Create(relativePath)
			fileContent, _ := ioutil.ReadFile(path)
			f.Write(fileContent)

			return nil
		})

	err = w.Close()
	if err != nil {
		fmt.Errorf("error while closing zip writer %v: ", err)
	}

	// create archive
	archiveFileName := fmt.Sprintf("%v_%v.zip", i.t.Name, i.t.versionTuple)
	archiveFilePath := fmt.Sprintf("%v/%v", outputPath, archiveFileName)
	a, err := os.Create(archiveFilePath)
	defer a.Close()

	if err != nil {
		return fmt.Errorf("Could not create file %v: ", archiveFilePath)
	}

	io.Copy(a, compressedBuffer)

	return nil
}
