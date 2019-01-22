/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

type MediaItem struct {
	MediaItem *types.MediaRecordType
	client    *Client
}

func NewMediaItem(cli *Client) *MediaItem {
	return &MediaItem{
		MediaItem: new(types.MediaRecordType),
		client:    cli,
	}
}

// Uploads an ISO file as media. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process. On upload fail client may need to
// remove vCD catalog item which waits for files to be uploaded.
func (vdc *Vdc) UploadMediaImage(mediaName, mediaDescription, filePath string, uploadPieceSize int64) (UploadTask, error) {
	util.Logger.Printf("[TRACE] UploadImage: %s, image name: %v \n", mediaName, mediaDescription)

	//	On a very high level the flow is as follows
	//	1. Makes a POST call to vCD to create media item(also creates a transfer folder in the spool area and as result will give a media item resource XML).
	//	2. Start uploading bits to the transfer folder
	//	3. Wait on the import task to finish on vCD side -> task success = upload complete

	if *vdc == (Vdc{}) {
		return UploadTask{}, errors.New("vdc can not be empty or nil")
	}

	mediaFilePath, err := validateAndFixFilePath(filePath)
	if err != nil {
		return UploadTask{}, err
	}

	isISOGood, err := verifyIso(mediaFilePath)
	if err != nil || !isISOGood {
		return UploadTask{}, fmt.Errorf("[ERROR] File %s isn't correct iso file: %#v", mediaFilePath, err)
	}

	mediaList, err := getExistingMediaItems(vdc)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Checking existing media files failed: %#v", err)
	}

	for _, media := range mediaList {
		if media.Name == mediaName {
			return UploadTask{}, fmt.Errorf("media item '%s' already exists. Upload with different name", mediaName)
		}
	}

	file, e := os.Stat(mediaFilePath)
	if e != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue finding file: %#v", e)
	}
	fileSize := file.Size()

	mediaItem, err := createMedia(vdc.client, vdc.Vdc.HREF+"/media", mediaName, mediaDescription, fileSize)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue creating media: %#v", err)
	}

	return executeUpload(vdc.client, mediaItem, mediaFilePath, mediaName, fileSize, uploadPieceSize)
}

func executeUpload(client *Client, mediaItem *types.Media, mediaFilePath, mediaName string, fileSize, uploadPieceSize int64) (UploadTask, error) {
	uploadLink, err := getUploadLink(mediaItem.Files)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue getting upload link: %#v", err)
	}

	callBack, uploadProgress := getCallBackFunction()

	uploadError := *new(error)

	details := uploadDetails{
		uploadLink:               uploadLink.String(), // just take string
		uploadedBytes:            0,
		fileSizeToUpload:         fileSize,
		uploadPieceSize:          uploadPieceSize,
		uploadedBytesForCallback: 0,
		allFilesSize:             fileSize,
		callBack:                 callBack,
		uploadError:              &uploadError,
	}

	go uploadFile(client, mediaFilePath, details)

	var task Task
	for _, item := range mediaItem.Tasks.Task {
		task, err = createTaskForVcdImport(client, item.HREF)
		if err != nil {
			removeImageOnError(client, mediaItem, mediaName)
			return UploadTask{}, err
		}
		if task.Task.Status == "error" {
			removeImageOnError(client, mediaItem, mediaName)
			return UploadTask{}, fmt.Errorf("task did not complete succesfully: %s", task.Task.Description)
		}
	}

	uploadTask := NewUploadTask(&task, uploadProgress, &uploadError)

	util.Logger.Printf("[TRACE] Upload media function finished and task for vcd import created. \n")

	return *uploadTask, nil
}

// Initiates creation of media item and returns temporary upload URL.
func createMedia(client *Client, link, mediaName, mediaDescription string, fileSize int64) (*types.Media, error) {
	uploadUrl, err := url.ParseRequestURI(link)
	if err != nil {
		return nil, fmt.Errorf("error getting vdc href: %v", err)
	}

	reqBody := bytes.NewBufferString(
		"<Media xmlns=\"http://www.vmware.com/vcloud/v1.5\" name=\"" + mediaName + "\" imageType=\"" + "iso" + "\" size=\"" + strconv.FormatInt(fileSize, 10) + "\" >" +
			"<Description>" + mediaDescription + "</Description>" +
			"</Media>")

	request := client.NewRequest(map[string]string{}, "POST", *uploadUrl, reqBody)
	request.Header.Add("Content-Type", "application/vnd.vmware.vcloud.media+xml")

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	mediaForUpload := &types.Media{}
	if err = decodeBody(response, mediaForUpload); err != nil {
		return nil, err
	}

	util.Logger.Printf("[TRACE] Media item parsed: %#v\n", mediaForUpload)

	if mediaForUpload.Tasks != nil {
		for _, task := range mediaForUpload.Tasks.Task {
			if "error" == task.Status && mediaName == mediaForUpload.Name {
				util.Logger.Printf("[Error] issue with creating media %#v", task.Error)
				return nil, fmt.Errorf("Error in vcd returned error code: %d, error: %s and message: %s ", task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
			}
		}
	}

	return mediaForUpload, nil
}

func removeImageOnError(client *Client, media *types.Media, itemName string) {
	if media != nil {
		util.Logger.Printf("[TRACE] Deleting media item %#v", media)

		// wait for task, cancel it and media item will be removed.
		var err error
		for {
			util.Logger.Printf("[TRACE] Sleep... for 5 seconds.\n")
			time.Sleep(time.Second * 5)
			media, err = queryMedia(client, media.HREF, itemName)
			if err != nil {
				util.Logger.Printf("[Error] Error deleting media item %v: %s", media, err)
			}
			if len(media.Tasks.Task) > 0 {
				util.Logger.Printf("[TRACE] Task found. Will try to cancel.\n")
				break
			}
		}

		for _, taskItem := range media.Tasks.Task {
			if itemName == taskItem.Owner.Name {
				task := NewTask(client)
				task.Task = taskItem
				err = task.CancelTask()
				if err != nil {
					util.Logger.Printf("[ERROR] Error canceling task for media upload %#v", err)
				}
			}
		}
	} else {
		util.Logger.Printf("[Error] Failed to delete media item created with error: %v", media)
	}
}

func queryMedia(client *Client, mediaUrl string, newItemName string) (*types.Media, error) {
	util.Logger.Printf("[TRACE] Querying media: %s\n", mediaUrl)

	parsedUrl, err := url.ParseRequestURI(mediaUrl)
	if err != nil {
		util.Logger.Printf("[Error] Error parsing url %v: %s", parsedUrl, err)
	}

	request := client.NewRequest(map[string]string{}, "GET", *parsedUrl, nil)
	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return nil, err
	}

	mediaParsed := &types.Media{}
	if err = decodeBody(response, mediaParsed); err != nil {
		return nil, err
	}

	defer response.Body.Close()

	for _, task := range mediaParsed.Tasks.Task {
		if "error" == task.Status && newItemName == task.Owner.Name {
			util.Logger.Printf("[Error] %#v", task.Error)
			return mediaParsed, fmt.Errorf("Error in vcd returned error code: %d, error: %s and message: %s ", task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
		}
	}

	return mediaParsed, nil
}

// Verifies provided file header matches standard
func verifyIso(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	return readHeader(file)
}

func readHeader(reader io.Reader) (bool, error) {
	buffer := make([]byte, 37000)

	_, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return false, err
	}

	headerOk := verifyHeader(buffer)

	if headerOk {
		return true, nil
	} else {
		return false, errors.New("file header didn't match ISO standard")
	}
}

// Verify file header info: https://www.garykessler.net/library/file_sigs.html
func verifyHeader(buf []byte) bool {
	// search for for CD001(43 44 30 30 31) in specific file places.
	//This signature usually occurs at byte offset 32769 (0x8001),
	//34817 (0x8801), or 36865 (0x9001).
	return (buf[32769] == 0x43 && buf[32770] == 0x44 &&
		buf[32771] == 0x30 && buf[32772] == 0x30 && buf[32773] == 0x31) ||
		(buf[34817] == 0x43 && buf[34818] == 0x44 &&
			buf[34819] == 0x30 && buf[34820] == 0x30 && buf[34821] == 0x31) ||
		(buf[36865] == 0x43 && buf[36866] == 0x44 &&
			buf[36867] == 0x30 && buf[36868] == 0x30 && buf[36869] == 0x31)
}

// Reference for API usage http://pubs.vmware.com/vcloud-api-1-5/wwhelp/wwhimpl/js/html/wwhelp.htm#href=api_prog/GUID-9356B99B-E414-474A-853C-1411692AF84C.html
// http://pubs.vmware.com/vcloud-api-1-5/wwhelp/wwhimpl/js/html/wwhelp.htm#href=api_prog/GUID-43DFF30E-391F-42DC-87B3-5923ABCEB366.html
func getExistingMediaItems(vdc *Vdc) ([]*types.MediaRecordType, error) {
	util.Logger.Printf("[TRACE] Querying medias \n")

	mediaResults, err := queryMediaItemsWithFilter(vdc, "vdc=="+url.QueryEscape(vdc.Vdc.HREF))

	util.Logger.Printf("[TRACE] Found media records: %d \n", len(mediaResults))
	return mediaResults, err
}

func queryMediaItemsWithFilter(vdc *Vdc, filter string) ([]*types.MediaRecordType, error) {
	typeMedia := "media"
	if vdc.client.IsSysAdmin {
		typeMedia = "adminMedia"
	}

	results, err := vdc.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia, "filter": filter})
	if err != nil {
		return nil, fmt.Errorf("error querying medias %#v", err)
	}

	mediaResults := results.Results.MediaRecord
	if vdc.client.IsSysAdmin {
		mediaResults = results.Results.AdminMediaRecord
	}
	return mediaResults, nil
}

// Looks for an Org Vdc network and, if found, will delete it.
func RemoveMediaImageIfExists(vdc Vdc, mediaName string) error {
	mediaItem, err := vdc.FindMediaImage(mediaName)
	if err == nil && mediaItem != (MediaItem{}) {
		task, err := mediaItem.Delete()
		if err != nil {
			return fmt.Errorf("error deleting media [phase 1] %s", mediaName)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error deleting media [task] %s", mediaName)
		}
	} else {
		util.Logger.Printf("[TRACE] Media not foun or error: %v - %#v \n", err, mediaItem)
	}
	return nil
}

// Deletes the Media Item, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Media.html
func (mediaItem *MediaItem) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] Deleting media item: %#v", mediaItem.MediaItem.Name)

	parsedUrl, err := url.ParseRequestURI(mediaItem.MediaItem.HREF)
	if err != nil {
		util.Logger.Printf("[Error] Error parsing url %v: %s", parsedUrl, err)
	}
	util.Logger.Printf("[TRACE] Url for deleting media item: %#v and name: %s", parsedUrl, mediaItem.MediaItem.Name)

	req := mediaItem.client.NewRequest(map[string]string{}, "DELETE", *parsedUrl, nil)

	resp, err := checkResp(mediaItem.client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error deleting Media item %s: %s", mediaItem.MediaItem.ID, err)
	}

	task := NewTask(mediaItem.client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil
}

// Finds media in catalog and returns catalog item
func FindMediaAsCatalogItem(org *Org, catalogName, mediaName string) (CatalogItem, error) {
	if catalogName == "" {
		return CatalogItem{}, errors.New("catalog name is empty")
	}
	if mediaName == "" {
		return CatalogItem{}, errors.New("media name is empty")
	}

	catalog, err := org.FindCatalog(catalogName)
	if err != nil || catalog == (Catalog{}) {
		return CatalogItem{}, fmt.Errorf("catalog not found or error %#v", err)
	}

	media, err := catalog.FindCatalogItem(mediaName)
	if err != nil || media == (CatalogItem{}) {
		return CatalogItem{}, fmt.Errorf("media not found or error %#v", err)
	}
	return media, nil
}
