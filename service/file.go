package service

import (
	"os"
	"path/filepath"
)

func CreateFilePath(filePath string) error {
	path, _ := filepath.Split(filePath)
	if len(path) == 0 {
		return nil
	}

	_, err := os.Stat(path)
	if err != nil || os.IsExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
	}
	return err
}

func OverwriteFolder(folderpath string) error {
	// If the directory exists, remove it
	if ok, _ := IsFolderExist(folderpath); !ok {
		err := os.RemoveAll(folderpath)
		if err != nil {
			return err
		}
	}

	// Create the directory
	if err := CreateFilePath(folderpath); err != nil {
		return err
	}

	return nil
}

func IsFolderExist(folderpath string) (bool, error) {
	if _, err := os.Stat(folderpath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
