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
	c *config
	t *target
}

func (i zipInstaller) createInstaller(outputPath string) error {
	fmt.Printf("Creating zip installer for target %v...\n", i.t.Name)
	var err error

	var compressedBuffer = new(bytes.Buffer)

	w := zip.NewWriter(compressedBuffer)

	artefactFolderPath := i.t.artefactFolderPath(i.c.Repo.Path)
	// compress everything in artefact path
	err = filepath.Walk(artefactFolderPath,
		func(path string, info os.FileInfo, err error) error {

			// do not archive folder...
			if info != nil && info.IsDir() {
				return nil
			}

			relativePath := strings.TrimPrefix(path, artefactFolderPath+"\\")
			fmt.Printf("Compressing %v | %v\n", path, relativePath)
			f, err := w.Create(relativePath)
			if err != nil {
				return fmt.Errorf("Error creating zip writer: %v\n", err)
			}

			fileContent, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("Error reading file: %v\n", err)
			}

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
