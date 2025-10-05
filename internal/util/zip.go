package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ZipDirectory(sourceDir, zipFileName string) error {
	file, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	isEmpty := true
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			_, err := zipWriter.Create(relPath + "/")
			return err
		}

		isEmpty = false

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		return err
	}

	if isEmpty {
		fmt.Printf("Warning: the directory is empty, the zip contains no files\n\n")
	}

	return nil
}
