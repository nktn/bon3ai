package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ClipboardType represents the type of clipboard operation
type ClipboardType int

const (
	ClipboardNone ClipboardType = iota
	ClipboardCopy
	ClipboardCut
)

// Clipboard holds copied/cut file paths
type Clipboard struct {
	Type  ClipboardType
	Paths []string
}

// Copy sets the clipboard to copy mode
func (c *Clipboard) Copy(paths []string) {
	c.Type = ClipboardCopy
	c.Paths = paths
}

// Cut sets the clipboard to cut mode
func (c *Clipboard) Cut(paths []string) {
	c.Type = ClipboardCut
	c.Paths = paths
}

// Clear clears the clipboard
func (c *Clipboard) Clear() {
	c.Type = ClipboardNone
	c.Paths = nil
}

// IsEmpty returns true if clipboard is empty
func (c *Clipboard) IsEmpty() bool {
	return c.Type == ClipboardNone || len(c.Paths) == 0
}

// CopyFile copies a file or directory to the destination directory
func CopyFile(src, destDir string) (string, error) {
	fileName := filepath.Base(src)
	dest := filepath.Join(destDir, fileName)
	dest = getUniquePath(dest)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	if srcInfo.IsDir() {
		err = copyDirRecursive(src, dest)
	} else {
		err = copyFileOnly(src, dest)
	}

	if err != nil {
		return "", err
	}
	return dest, nil
}

// MoveFile moves a file or directory to the destination directory
func MoveFile(src, destDir string) (string, error) {
	fileName := filepath.Base(src)
	dest := filepath.Join(destDir, fileName)
	dest = getUniquePath(dest)

	// Try simple rename first
	err := os.Rename(src, dest)
	if err == nil {
		return dest, nil
	}

	// If rename fails (cross-device), copy then delete
	srcInfo, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	if srcInfo.IsDir() {
		if err := copyDirRecursive(src, dest); err != nil {
			return "", err
		}
		if err := os.RemoveAll(src); err != nil {
			return "", err
		}
	} else {
		if err := copyFileOnly(src, dest); err != nil {
			return "", err
		}
		if err := os.Remove(src); err != nil {
			return "", err
		}
	}

	return dest, nil
}

// DeleteFile deletes a file or directory
func DeleteFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

// RenameFile renames a file or directory
func RenameFile(path, newName string) (string, error) {
	parent := filepath.Dir(path)
	newPath := filepath.Join(parent, newName)

	if newPath == path {
		return path, nil
	}

	if _, err := os.Stat(newPath); err == nil {
		return "", fmt.Errorf("file already exists: %s", newPath)
	}

	if err := os.Rename(path, newPath); err != nil {
		return "", err
	}

	return newPath, nil
}

// CreateFile creates a new empty file
func CreateFile(parentDir, name string) (string, error) {
	path := filepath.Join(parentDir, name)

	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("file already exists: %s", path)
	}

	file, err := os.Create(path)
	if err != nil {
		return "", err
	}
	file.Close()

	return path, nil
}

// CreateDirectory creates a new directory
func CreateDirectory(parentDir, name string) (string, error) {
	path := filepath.Join(parentDir, name)

	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("directory already exists: %s", path)
	}

	if err := os.Mkdir(path, 0755); err != nil {
		return "", err
	}

	return path, nil
}

// copyFileOnly copies a single file
func copyFileOnly(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// copyDirRecursive copies a directory recursively
func copyDirRecursive(src, dest string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFileOnly(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// getUniquePath returns a unique path by appending _N suffix if needed
func getUniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]

	counter := 1
	for {
		newPath := fmt.Sprintf("%s_%d%s", base, counter, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}
