/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

// uploadLink - vCD created temporary upload link
// uploadedBytes - how much of file already uploaded
// fileSizeToUpload - how much bytes will be uploaded
// uploadPieceSize - size of chunks in which the file will be uploaded to the catalog.
// uploadedBytesForCallback all uploaded bytes if multi disk in ova
// allFilesSize overall sum of size if multi disk in ova
// callBack a function with signature //function(bytesUpload, totalSize) to let the caller monitor progress of the upload operation.
type uploadDetails struct {
	uploadLink                                                                               string
	uploadedBytes, fileSizeToUpload, uploadPieceSize, uploadedBytesForCallback, allFilesSize int64
	callBack                                                                                 func(bytesUpload, totalSize int64)
	uploadError                                                                              *error
}

// Upload file by parts which size is defined by user provided variable uploadPieceSize and
// provides how much bytes uploaded to callback. Callback allows to monitor upload progress.
// params:
// client - client for requests
// filePath - file path to file which will be uploaded
// uploadDetails - file upload settings and data
func uploadFile(client *Client, filePath string, uDetails uploadDetails) (int64, error) {
	util.Logger.Printf("[TRACE] Starting uploading: %s, offset: %v, fileze: %v, toLink: %s \n", filePath, uDetails.uploadedBytes, uDetails.fileSizeToUpload, uDetails.uploadLink)

	var part []byte
	var count int
	var pieceSize int64

	// do not allow smaller than 1kb
	if uDetails.uploadPieceSize > 1024 && uDetails.uploadPieceSize < uDetails.fileSizeToUpload {
		pieceSize = uDetails.uploadPieceSize
	} else {
		pieceSize = defaultPieceSize
	}

	util.Logger.Printf("[TRACE] Uploading will use piece size: %#v \n", pieceSize)

	file, err := os.Open(filePath)
	if err != nil {
		util.Logger.Printf("[ERROR] during upload process - file open issue : %s, error %#v ", filePath, err)
		*uDetails.uploadError = err
		return 0, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		util.Logger.Printf("[ERROR] during upload process - file issue : %s, error %#v ", filePath, err)
		*uDetails.uploadError = err
		return 0, err
	}

	defer file.Close()

	part = make([]byte, pieceSize)

	for {
		if count, err = io.ReadFull(file, part); err != nil {
			break
		}
		err = uploadPartFile(client, part, int64(count), uDetails)
		uDetails.uploadedBytes += int64(count)
		uDetails.uploadedBytesForCallback += int64(count)
		if err != nil {
			util.Logger.Printf("[ERROR] during upload process: %s, error %#v ", filePath, err)
			*uDetails.uploadError = err
			return 0, err
		}
	}

	// upload last part as ReadFull returns io.ErrUnexpectedEOF when reaches end of file.
	if err == io.ErrUnexpectedEOF {
		err = uploadPartFile(client, part[:count], int64(count), uDetails)
		if err != nil {
			util.Logger.Printf("[ERROR] during upload process: %s, error %#v ", filePath, err)
			*uDetails.uploadError = err
			return 0, err
		}
	} else {
		util.Logger.Printf("Error Uploading: %s, error %#v ", filePath, err)
		*uDetails.uploadError = err
		return 0, err
	}

	return fileInfo.Size(), nil
}

// Create Request with right headers and range settings. Support multi part file upload.
// requestUrl - upload url
// filePart - bytes to upload
// offset - how much is uploaded
// filePartSize - how much bytes will be uploaded
// fileSizeToUpload - final file size
func newFileUploadRequest(requestUrl string, filePart []byte, offset, filePartSize, fileSizeToUpload int64) (*http.Request, error) {
	util.Logger.Printf("[TRACE] Creating file upload request: %s, %v, %v, %v \n", requestUrl, offset, filePartSize, fileSizeToUpload)

	uploadReq, err := http.NewRequest("PUT", requestUrl, bytes.NewReader(filePart))
	if err != nil {
		return nil, err
	}

	uploadReq.ContentLength = filePartSize
	uploadReq.Header.Set("Content-Length", strconv.FormatInt(uploadReq.ContentLength, 10))

	rangeExpression := "bytes " + strconv.FormatInt(int64(offset), 10) + "-" + strconv.FormatInt(int64(offset+filePartSize-1), 10) + "/" + strconv.FormatInt(int64(fileSizeToUpload), 10)
	uploadReq.Header.Set("Content-Range", rangeExpression)

	for key, value := range uploadReq.Header {
		util.Logger.Printf("[TRACE] Header: %s :%s \n", key, value)
	}

	return uploadReq, nil
}

// Initiates file part upload by creating request and running it.
// params:
// client - client for requests
// part - bytes of file part
// partDataSize - how much bytes will be uploaded
// uploadDetails - file upload settings and data
func uploadPartFile(client *Client, part []byte, partDataSize int64, uDetails uploadDetails) error {
	request, err := newFileUploadRequest(uDetails.uploadLink, part, uDetails.uploadedBytes, partDataSize, uDetails.fileSizeToUpload)
	if err != nil {
		return err
	}

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return fmt.Errorf("File upload failed. Err: %s \n", err)
	}
	response.Body.Close()

	uDetails.callBack(uDetails.uploadedBytesForCallback+partDataSize, uDetails.allFilesSize)

	return nil
}

func getUploadLink(files *types.FilesList) (*url.URL, error) {
	util.Logger.Printf("[TRACE] getUploadLink - Parsing upload link: %#v\n", files)

	if len(files.File) > 1 {
		return nil, errors.New("unexpected response from vCD: found more than one link for upload")
	}

	ovfUploadHref, err := url.ParseRequestURI(files.File[0].Link[0].HREF)
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("[TRACE] getUploadLink- upload link found: %#v\n", ovfUploadHref)
	return ovfUploadHref, nil
}

func createTaskForVcdImport(client *Client, taskHREF string) (Task, error) {
	util.Logger.Printf("[TRACE] Create task for vcd with HREF: %s\n", taskHREF)

	taskURL, err := url.ParseRequestURI(taskHREF)
	if err != nil {
		return Task{}, err
	}

	request := client.NewRequest(map[string]string{}, "GET", *taskURL, nil)
	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return Task{}, err
	}

	task := NewTask(client)

	if err = decodeBody(response, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

func getCallBackFunction() (func(int64, int64), *float64) {
	var uploadProgress float64
	callback := func(bytesUploaded, totalSize int64) {
		uploadProgress = (float64(bytesUploaded) / float64(totalSize)) * 100
	}
	return callback, &uploadProgress
}

func validateAndFixFilePath(file string) (string, error) {
	absolutePath, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	fileInfo, err := os.Stat(absolutePath)
	if os.IsNotExist(err) {
		return "", err
	}
	if fileInfo.Size() == 0 {
		return "", errors.New("file is empty")
	}
	return absolutePath, nil
}
