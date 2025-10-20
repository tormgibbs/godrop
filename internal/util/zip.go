package util

import (
	"archive/zip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var (
	nouns      = []string{"fox", "owl", "star", "cloud", "tree", "river", "moon", "sun", "rock", "wave", "breeze", "shard", "gem", "mist", "dawn", "dusk", "flame", "peak", "canyon", "echo"}
	adjectives = []string{"quick", "silent", "bright", "dark", "clear", "calm", "swift", "deep", "soft", "loud", "fresh", "frozen", "warm", "shining", "gentle", "smooth", "rough", "hollow", "solid", "vivid"}
	r          = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func ZipDirectory(sourceDir, destZip string) error {
	paths := []string{sourceDir}

	count, err := ZipPaths(paths, destZip)
	if err != nil {
		return err
	}

	if count == 0 {
		fmt.Printf("Warning: the directory is empty, the zip contains no files\n\n")
	}

	return nil
}

func ZipPaths(sourcePaths []string, destZip string) (int, error) {
	outFile, err := os.Create(destZip)
	if err != nil {
		return 0, fmt.Errorf("failed to create zip file %q: %w", destZip, err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	var fileCount int

	for _, sourcePath := range sourcePaths {
		sourcePath = filepath.Clean(sourcePath)
		info, err := os.Stat(sourcePath)
		if err != nil {
			return fileCount, fmt.Errorf("error stating source path %q: %w", sourcePath, err)
		}

		if info.IsDir() {
			if err := addDirectoryToZip(zipWriter, sourcePath, info.Name(), &fileCount); err != nil {
				return fileCount, err
			}
		} else {
			if err := addFileToZip(zipWriter, sourcePath, info.Name(), &fileCount); err != nil {
				return fileCount, err
			}
		}
	}

	if err := zipWriter.Close(); err != nil {
		return fileCount, fmt.Errorf("error closing zip writer: %w", err)
	}

	return fileCount, nil
}

func addDirectoryToZip(zw *zip.Writer, sourceDir, baseDirInZip string, fileCount *int) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(baseDirInZip, relPath)

		if info.IsDir() {
			if relPath != "" {
				_, err = zw.CreateHeader(&zip.FileHeader{
					Name:     zipPath + "/",
					Method:   zip.Deflate,
					Modified: info.ModTime(),
				})
				if err != nil {
					return fmt.Errorf("failed to create directory header for %q: %w", zipPath, err)
				}
			}
		} else {
			if err := addFileToZip(zw, path, zipPath, fileCount); err != nil {
				return err
			}
		}
		return nil
	})
}

func addFileToZip(zw *zip.Writer, filePath, zipPath string, fileCount *int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file %q: %w", filePath, err)
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("failed to create file header for %q: %w", filePath, err)
	}

	header.Name = zipPath
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip writer for %q: %w", filePath, err)
	}

	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("failed to copy file %q content: %w", filePath, err)
	}

	*fileCount++

	return nil
}

func GenerateRandomName() string {
	adj := adjectives[r.Intn(len(adjectives))]
	noun := nouns[r.Intn(len(nouns))]
	return fmt.Sprintf("%s-%s", adj, noun)
}
