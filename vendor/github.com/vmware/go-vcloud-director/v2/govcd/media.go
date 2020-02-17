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
	"strconv"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Deprecated: use MediaRecord
type MediaItem struct {
	MediaItem *types.MediaRecordType
	vdc       *Vdc
}

// Deprecated: use NewMediaRecord
func NewMediaItem(vdc *Vdc) *MediaItem {
	return &MediaItem{
		MediaItem: new(types.MediaRecordType),
		vdc:       vdc,
	}
}

type Media struct {
	Media  *types.Media
	client *Client
}

func NewMedia(cli *Client) *Media {
	return &Media{
		Media:  new(types.Media),
		client: cli,
	}
}

type MediaRecord struct {
	MediaRecord *types.MediaRecordType
	client      *Client
}

func NewMediaRecord(cli *Client) *MediaRecord {
	return &MediaRecord{
		MediaRecord: new(types.MediaRecordType),
		client:      cli,
	}
}

// Uploads an ISO file as media. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process. On upload fail client may need to
// remove vCD catalog item which waits for files to be uploaded.
// Recommend to use catalog.UploadMediaImage which let's expect in which catalog file will be placed
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
		return UploadTask{}, fmt.Errorf("[ERROR] File %s isn't correct iso file: %s", mediaFilePath, err)
	}

	mediaList, err := getExistingMedia(vdc)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Checking existing media files failed: %s", err)
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

	media, err := createMedia(vdc.client, vdc.Vdc.HREF+"/media", mediaName, mediaDescription, fileSize)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue creating media: %s", err)
	}

	return executeUpload(vdc.client, media, mediaFilePath, mediaName, fileSize, uploadPieceSize)
}

func executeUpload(client *Client, media *types.Media, mediaFilePath, mediaName string, fileSize, uploadPieceSize int64) (UploadTask, error) {
	uploadLink, err := getUploadLink(media.Files)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue getting upload link: %s", err)
	}

	callBack, uploadProgress := getProgressCallBackFunction()

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
	for _, item := range media.Tasks.Task {
		task, err = createTaskForVcdImport(client, item.HREF)
		if err != nil {
			removeImageOnError(client, media, mediaName)
			return UploadTask{}, err
		}
		if task.Task.Status == "error" {
			removeImageOnError(client, media, mediaName)
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
		return nil, fmt.Errorf("error getting vdc href: %s", err)
	}

	reqBody := bytes.NewBufferString(
		"<Media xmlns=\"" + types.XMLNamespaceVCloud + "\" name=\"" + mediaName + "\" imageType=\"" + "iso" + "\" size=\"" + strconv.FormatInt(fileSize, 10) + "\" >" +
			"<Description>" + mediaDescription + "</Description>" +
			"</Media>")

	request := client.NewRequest(map[string]string{}, http.MethodPost, *uploadUrl, reqBody)
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
			if task.Status == "error" && mediaName == mediaForUpload.Name {
				util.Logger.Printf("[Error] issue with creating media %#v", task.Error)
				return nil, fmt.Errorf("error in vcd returned error code: %d, error: %s and message: %s ", task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
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
					util.Logger.Printf("[ERROR] Error canceling task for media upload %s", err)
				}
			}
		}
	} else {
		util.Logger.Printf("[Error] Failed to delete media item created with error: %v", media)
	}
}

func queryMedia(client *Client, mediaUrl string, newItemName string) (*types.Media, error) {
	util.Logger.Printf("[TRACE] Querying media: %s\n", mediaUrl)

	mediaParsed := &types.Media{}

	_, err := client.ExecuteRequest(mediaUrl, http.MethodGet,
		"", "error quering media: %s", nil, mediaParsed)
	if err != nil {
		return nil, err
	}

	for _, task := range mediaParsed.Tasks.Task {
		if task.Status == "error" && newItemName == task.Owner.Name {
			util.Logger.Printf("[Error] %#v", task.Error)
			return mediaParsed, fmt.Errorf("error in vcd returned error code: %d, error: %s and message: %s ", task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
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
	// search for CD001(43 44 30 30 31) in specific file places.
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
func getExistingMedia(vdc *Vdc) ([]*types.MediaRecordType, error) {
	util.Logger.Printf("[TRACE] Querying medias \n")

	mediaResults, err := queryMediaWithFilter(vdc, "vdc=="+url.QueryEscape(vdc.Vdc.HREF))

	util.Logger.Printf("[TRACE] Found media records: %d \n", len(mediaResults))
	return mediaResults, err
}

func queryMediaWithFilter(vdc *Vdc, filter string) ([]*types.MediaRecordType, error) {
	typeMedia := "media"
	if vdc.client.IsSysAdmin {
		typeMedia = "adminMedia"
	}

	results, err := vdc.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia, "filter": filter})
	if err != nil {
		return nil, fmt.Errorf("error querying medias %s", err)
	}

	mediaResults := results.Results.MediaRecord
	if vdc.client.IsSysAdmin {
		mediaResults = results.Results.AdminMediaRecord
	}
	return mediaResults, nil
}

// Looks for media and, if found, will delete it.
// Deprecated: Use catalog.RemoveMediaIfExist
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
		util.Logger.Printf("[TRACE] Media not found or error: %s - %#v \n", err, mediaItem)
	}
	return nil
}

// Looks for media and, if found, will delete it.
func (adminCatalog *AdminCatalog) RemoveMediaIfExists(mediaName string) error {
	media, err := adminCatalog.GetMediaByName(mediaName, true)
	if err == nil {
		task, err := media.Delete()
		if err != nil {
			return fmt.Errorf("error deleting media [phase 1] %s", mediaName)
		}
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error deleting media [task] %s", mediaName)
		}
	} else {
		util.Logger.Printf("[TRACE] Media not found or error: %s - %#v \n", err, media)
	}
	return nil
}

// Deletes the Media Item, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Media.html
// Deprecated: Use MediaRecord.Delete
func (mediaItem *MediaItem) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] Deleting media item: %#v", mediaItem.MediaItem.Name)

	// Return the task
	return mediaItem.vdc.client.ExecuteTaskRequest(mediaItem.MediaItem.HREF, http.MethodDelete,
		"", "error deleting Media item: %s", nil)
}

// Deletes the Media Item, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Media.html
func (media *Media) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] Deleting media item: %#v", media.Media.Name)

	// Return the task
	return media.client.ExecuteTaskRequest(media.Media.HREF, http.MethodDelete,
		"", "error deleting Media item: %s", nil)
}

// Finds media in catalog and returns catalog item
// Deprecated: Use catalog.GetMediaByName()
func FindMediaAsCatalogItem(org *Org, catalogName, mediaName string) (CatalogItem, error) {
	if catalogName == "" {
		return CatalogItem{}, errors.New("catalog name is empty")
	}
	if mediaName == "" {
		return CatalogItem{}, errors.New("media name is empty")
	}

	catalog, err := org.FindCatalog(catalogName)
	if err != nil || catalog == (Catalog{}) {
		return CatalogItem{}, fmt.Errorf("catalog not found or error %s", err)
	}

	media, err := catalog.FindCatalogItem(mediaName)
	if err != nil || media == (CatalogItem{}) {
		return CatalogItem{}, fmt.Errorf("media not found or error %s", err)
	}
	return media, nil
}

// Refresh refreshes the media item information by href
// Deprecated: Use MediaRecord.Refresh
func (mediaItem *MediaItem) Refresh() error {

	if mediaItem.MediaItem == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	if mediaItem.MediaItem.Name == "nil" {
		return fmt.Errorf("cannot refresh, Name is empty")
	}

	latestMediaItem, err := mediaItem.vdc.FindMediaImage(mediaItem.MediaItem.Name)
	*mediaItem = latestMediaItem

	return err
}

// Refresh refreshes the media information by href
func (media *Media) Refresh() error {

	if media.Media == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	url := media.Media.HREF

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	media.Media = &types.Media{}

	_, err := media.client.ExecuteRequest(url, http.MethodGet,
		"", "error retrieving media: %s", nil, media.Media)

	return err
}

// GetMediaByHref finds a Media by HREF
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetMediaByHref(mediaHref string) (*Media, error) {

	media := NewMedia(cat.client)

	_, err := cat.client.ExecuteRequest(mediaHref, http.MethodGet,
		"", "error retrieving media: %#v", nil, media.Media)
	if err != nil && strings.Contains(err.Error(), "MajorErrorCode:403") {
		return nil, ErrorEntityNotFound
	}
	if err != nil {
		return nil, err
	}
	return media, nil
}

// GetMediaByName finds a Media by Name
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetMediaByName(mediaName string, refresh bool) (*Media, error) {
	if refresh {
		err := cat.Refresh()
		if err != nil {
			return nil, err
		}
	}
	for _, catalogItems := range cat.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == mediaName && catalogItem.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				catalogItemElement, err := cat.GetCatalogItemByHref(catalogItem.HREF)
				if err != nil {
					return nil, err
				}
				return cat.GetMediaByHref(catalogItemElement.CatalogItem.Entity.HREF)
			}
		}
	}
	return nil, ErrorEntityNotFound
}

// GetMediaById finds a Media by ID
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (catalog *Catalog) GetMediaById(mediaId string) (*Media, error) {
	typeMedia := "media"
	if catalog.client.IsSysAdmin {
		typeMedia = "adminMedia"
	}

	results, err := catalog.client.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia,
		"filter": fmt.Sprintf("catalogName==%s", url.QueryEscape(catalog.Catalog.Name))})
	if err != nil {
		return nil, fmt.Errorf("error querying medias %s", err)
	}

	mediaResults := results.Results.MediaRecord
	if catalog.client.IsSysAdmin {
		mediaResults = results.Results.AdminMediaRecord
	}
	for _, mediaRecord := range mediaResults {
		if equalIds(mediaId, mediaRecord.ID, mediaRecord.HREF) {
			return catalog.GetMediaByHref(mediaRecord.HREF)
		}
	}
	return nil, ErrorEntityNotFound
}

// GetMediaByNameOrId finds a Media by Name or ID
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (cat *Catalog) GetMediaByNameOrId(identifier string, refresh bool) (*Media, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return cat.GetMediaByName(name, refresh) }
	getById := func(id string, refresh bool) (interface{}, error) { return cat.GetMediaById(id) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
	if entity == nil {
		return nil, err
	}
	return entity.(*Media), err
}

// GetMediaByHref finds a Media by HREF
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (adminCatalog *AdminCatalog) GetMediaByHref(mediaHref string) (*Media, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.GetMediaByHref(mediaHref)
}

// GetMediaByName finds a Media by Name
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (adminCatalog *AdminCatalog) GetMediaByName(mediaName string, refresh bool) (*Media, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.GetMediaByName(mediaName, refresh)
}

// GetMediaById finds a Media by ID
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (adminCatalog *AdminCatalog) GetMediaById(mediaId string) (*Media, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.GetMediaById(mediaId)
}

// GetMediaByNameOrId finds a Media by Name or ID
// On success, returns a pointer to the Media structure and a nil error
// On failure, returns a nil pointer and an error
func (adminCatalog *AdminCatalog) GetMediaByNameOrId(identifier string, refresh bool) (*Media, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.GetMediaByNameOrId(identifier, refresh)
}

// QueryMedia returns media image found in system using `name` and `catalog name` as query.
func (catalog *Catalog) QueryMedia(mediaName string) (*MediaRecord, error) {
	util.Logger.Printf("[TRACE] Querying medias by name and catalog\n")

	if catalog == nil || catalog.Catalog == nil || catalog.Catalog.Name == "" {
		return nil, errors.New("catalog is empty")
	}
	if mediaName == "" {
		return nil, errors.New("media name is empty")
	}

	typeMedia := "media"
	if catalog.client.IsSysAdmin {
		typeMedia = "adminMedia"
	}

	results, err := catalog.client.QueryWithNotEncodedParams(nil, map[string]string{"type": typeMedia,
		"filter": fmt.Sprintf("name==%s;catalogName==%s",
			url.QueryEscape(mediaName),
			url.QueryEscape(catalog.Catalog.Name))})
	if err != nil {
		return nil, fmt.Errorf("error querying medias %s", err)
	}
	newMediaRecord := NewMediaRecord(catalog.client)

	mediaResults := results.Results.MediaRecord
	if catalog.client.IsSysAdmin {
		mediaResults = results.Results.AdminMediaRecord
	}
	if len(mediaResults) == 1 {
		newMediaRecord.MediaRecord = mediaResults[0]
	}

	if len(mediaResults) == 0 {
		return nil, ErrorEntityNotFound
	}
	// this shouldn't happen, but we will check anyways
	if len(mediaResults) > 1 {
		return nil, fmt.Errorf("found more than one result %#v with catalog name %s and media name %s ", mediaResults, catalog.Catalog.Name, mediaName)
	}

	util.Logger.Printf("[TRACE] Found media record by name: %#v \n", mediaResults[0])
	return newMediaRecord, nil
}

// QueryMedia returns media image found in system using `name` and `catalog name` as query.
func (adminCatalog *AdminCatalog) QueryMedia(mediaName string) (*MediaRecord, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.QueryMedia(mediaName)
}

// Refresh refreshes the media information by href
func (mediaRecord *MediaRecord) Refresh() error {

	if mediaRecord.MediaRecord == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	if mediaRecord.MediaRecord.Name == "" {
		return fmt.Errorf("cannot refresh, Name is empty")
	}

	url := mediaRecord.MediaRecord.HREF

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	mediaRecord.MediaRecord = &types.MediaRecordType{}

	_, err := mediaRecord.client.ExecuteRequest(url, http.MethodGet,
		"", "error retrieving media: %s", nil, mediaRecord.MediaRecord)

	return err
}

// Deletes the Media Item, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Media.html
func (mediaRecord *MediaRecord) Delete() (Task, error) {
	util.Logger.Printf("[TRACE] Deleting media item: %#v", mediaRecord.MediaRecord.Name)

	// Return the task
	return mediaRecord.client.ExecuteTaskRequest(mediaRecord.MediaRecord.HREF, http.MethodDelete,
		"", "error deleting Media item: %s", nil)
}
