/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package util

import (
	"archive/tar"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const TmpDirPrefix = "govcd"

// Extract files to system tmp dir with name govcd+random number. Created folder with files isn't deleted.
// Returns extracted files paths in array and path where folder with files created.
func Unpack(tarFile string) ([]string, string, error) {

	var filePaths []string
	var dst string

	reader, err := os.Open(tarFile)
	if err != nil {
		return filePaths, dst, err
	}
	defer reader.Close()

	tarReader := tar.NewReader(reader)

	dst, err = ioutil.TempDir("", TmpDirPrefix)
	if err != nil {
		return filePaths, dst, err
	}

	var expectedFileSize int64 = -1

	for {
		header, err := tarReader.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return filePaths, dst, nil

			// return any other error
		case err != nil:
			return filePaths, dst, err

			// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue

		case header != nil:
			expectedFileSize = header.Size
		}

		// the target location where the dir/newFile should be created
		target := filepath.Join(dst, sanitizedName(header.Name))
		Logger.Printf("[TRACE] extracting newFile: %s \n", target)

		// check the newFile type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return filePaths, dst, err
				}
			}

		case tar.TypeSymlink:
			if header.Linkname != "" {
				err := os.Symlink(header.Linkname, target)
				if err != nil {
					return filePaths, dst, err
				}
			} else {
				return filePaths, dst, errors.New("file %s is a symlink, but no link information was provided")
			}

			// if it's a newFile create it
		case tar.TypeReg:
			newFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return filePaths, dst, err
			}

			// copy over contents
			if _, err := io.Copy(newFile, tarReader); err != nil {
				return filePaths, dst, err
			}

			filePaths = append(filePaths, newFile.Name())

			if err := isExtractedFileValid(newFile, expectedFileSize); err != nil {
				newFile.Close()
				return filePaths, dst, err
			}

			// manually close here after each newFile operation; defering would cause each newFile close
			// to wait until all operations have completed.
			newFile.Close()
		}
	}
}

func isExtractedFileValid(file *os.File, expectedFileSize int64) error {
	if fInfo, err := file.Stat(); err == nil {
		Logger.Printf("[TRACE] isExtractedFileValid: created file size %#v, size from header %#v.\n", fInfo.Size(), expectedFileSize)
		if fInfo.Size() != expectedFileSize && expectedFileSize != -1 {
			return errors.New("extracted file didn't match defined file size")
		}
	}
	return nil
}

func sanitizedName(filename string) string {
	if len(filename) > 1 && filename[1] == ':' {
		filename = filename[2:]
	}
	filename = strings.TrimLeft(filename, "\\/.")
	filename = strings.TrimLeft(filename, "./")
	filename = strings.Replace(filename, "../../", "../", -1)
	return strings.Replace(filename, "..\\", "", -1)
}
