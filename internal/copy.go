package oda

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// From: https://stackoverflow.com/questions/51779243/copy-a-folder-in-go

func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return fmt.Errorf("failed to list directory: '%s', error: '%s'", scrDir, err.Error())
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to get file info for '%s'", sourcePath)
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory: '%s', error: '%s'", destPath, err.Error())
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy directory: '%s', error: '%s'", sourcePath, err.Error())
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy symlink: '%s', error: '%s'", sourcePath, err.Error())
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return fmt.Errorf("failed to copy file: '%s', error: '%s'", sourcePath, err.Error())
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return fmt.Errorf("failed to change owner for '%s'", destPath)
		}

		fInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info for '%s'", sourcePath)
		}

		isSymlink := fInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fInfo.Mode()); err != nil {
				return fmt.Errorf("failed to change mode for '%s'", destPath)
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("failed to create file: '%s', error: '%s'", dstFile, err.Error())
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open file: '%s', error: '%s'", srcFile, err.Error())
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("failed to copy file: '%s', error: '%s'", srcFile, err.Error())
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return fmt.Errorf("failed to read symlink: '%s', error: '%s'", source, err.Error())
	}
	return os.Symlink(link, dest)
}
