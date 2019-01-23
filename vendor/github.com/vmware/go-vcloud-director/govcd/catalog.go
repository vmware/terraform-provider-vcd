/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

const (
	defaultPieceSize int64 = 1024 * 1024
)

type CatalogOperations interface {
	FindCatalogItem(catalogItem string) (CatalogItem, error)
}

// AdminCatalog is a admin view of a vCloud Director Catalog
// To be able to get an AdminCatalog representation, users must have
// admin credentials to the System org. AdminCatalog is used
// for creating, updating, and deleting a Catalog.
// Definition: https://code.vmware.com/apis/220/vcloud#/doc/doc/types/AdminCatalogType.html
type AdminCatalog struct {
	AdminCatalog *types.AdminCatalog
	client       *Client
}

type Catalog struct {
	Catalog *types.Catalog
	client  *Client
}

func NewCatalog(client *Client) *Catalog {
	return &Catalog{
		Catalog: new(types.Catalog),
		client:  client,
	}
}

func NewAdminCatalog(client *Client) *AdminCatalog {
	return &AdminCatalog{
		AdminCatalog: new(types.AdminCatalog),
		client:       client,
	}
}

// Deletes the Catalog, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Catalog.html
func (catalog *Catalog) Delete(force, recursive bool) error {

	adminCatalogHREF := catalog.client.VCDHREF
	adminCatalogHREF.Path += "/admin/catalog/" + getEntityNumericId(catalog.Catalog.ID)

	req := catalog.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, "DELETE", adminCatalogHREF, nil)

	_, err := checkResp(catalog.client.Http.Do(req))

	if err != nil {
		return fmt.Errorf("error deleting Catalog %s: %s", catalog.Catalog.ID, err)
	}

	return nil
}

// slice only numeric part from ID:"urn:vcloud:catalog:97384890-180c-4563-b9b7-0dc50a2430b0"
func getEntityNumericId(catalogId string) string {
	if catalogId != "" && (len(catalogId) > 19) {
		return catalogId[19:]
	}
	return ""
}

// Deletes the Catalog, returning an error if the vCD call fails.
// Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/DELETE-Catalog.html
func (adminCatalog *AdminCatalog) Delete(force, recursive bool) error {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.Delete(force, recursive)
}

//   Updates the Catalog definition from current Catalog struct contents.
//   Any differences that may be legally applied will be updated.
//   Returns an error if the call to vCD fails. Update automatically performs
//   a refresh with the admin catalog it gets back from the rest api
//   Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/PUT-Catalog.html
func (adminCatalog *AdminCatalog) Update() error {
	reqCatalog := &types.Catalog{
		Name:        adminCatalog.AdminCatalog.Catalog.Name,
		Description: adminCatalog.AdminCatalog.Description,
	}
	vcomp := &types.AdminCatalog{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Catalog:     *reqCatalog,
		IsPublished: adminCatalog.AdminCatalog.IsPublished,
	}
	adminCatalogHREF, err := url.ParseRequestURI(adminCatalog.AdminCatalog.HREF)
	if err != nil {
		return fmt.Errorf("error parsing admin catalog's href: %v", err)
	}
	output, err := xml.MarshalIndent(vcomp, "  ", "    ")
	if err != nil {
		return fmt.Errorf("error marshalling xml data for update %v", err)
	}
	xmlData := bytes.NewBufferString(xml.Header + string(output))
	req := adminCatalog.client.NewRequest(map[string]string{}, "PUT", *adminCatalogHREF, xmlData)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.catalog+xml")
	resp, err := checkResp(adminCatalog.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error updating catalog: %s : %s", err, adminCatalogHREF.Path)
	}

	catalog := &types.AdminCatalog{}
	if err = decodeBody(resp, catalog); err != nil {
		return fmt.Errorf("error decoding update response: %s", err)
	}
	adminCatalog.AdminCatalog = catalog
	return nil
}

// Envelope is a ovf description root element. File contains information for vmdk files.
// Namespace: http://schemas.dmtf.org/ovf/envelope/1
// Description: Envelope is a ovf description root element. File contains information for vmdk files..
type Envelope struct {
	File []struct {
		HREF      string `xml:"href,attr"`
		ID        string `xml:"id,attr"`
		Size      int    `xml:"size,attr"`
		ChunkSize int    `xml:"chunkSize,attr"`
	} `xml:"References>File"`
}

// If catalog item is a valid CatalogItem and the call succeeds,
// then the function returns a CatalogItem. If the item does not
// exist, then it returns an empty CatalogItem. If the call fails
// at any point, it returns an error.
func (cat *Catalog) FindCatalogItem(catalogItemName string) (CatalogItem, error) {
	for _, catalogItems := range cat.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == catalogItemName && catalogItem.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				catalogItemHREF, err := url.ParseRequestURI(catalogItem.HREF)

				if err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				req := cat.client.NewRequest(map[string]string{}, "GET", *catalogItemHREF, nil)

				resp, err := checkResp(cat.client.Http.Do(req))
				if err != nil {
					return CatalogItem{}, fmt.Errorf("error retrieving catalog: %s", err)
				}

				cat := NewCatalogItem(cat.client)

				if err = decodeBody(resp, cat.CatalogItem); err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				// The request was successful
				return *cat, nil
			}
		}
	}

	return CatalogItem{}, nil
}

// Uploads an ova file to a catalog. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process. On upload fail client may need to
// remove vCD catalog item which waits for files to be uploaded. Files from ova are extracted to system
// temp folder "govcd+random number" and left for inspection on error.
func (adminCatalog *AdminCatalog) UploadOvf(ovaFileName, itemName, description string, uploadPieceSize int64) (UploadTask, error) {
	catalog := NewCatalog(adminCatalog.client)
	catalog.Catalog = &adminCatalog.AdminCatalog.Catalog
	return catalog.UploadOvf(ovaFileName, itemName, description, uploadPieceSize)
}

// Uploads an ova file to a catalog. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process. On upload fail client may need to
// remove vCD catalog item which waits for files to be uploaded. Files from ova are extracted to system
// temp folder "govcd+random number" and left for inspection on error.
func (cat *Catalog) UploadOvf(ovaFileName, itemName, description string, uploadPieceSize int64) (UploadTask, error) {

	//	On a very high level the flow is as follows
	//	1. Makes a POST call to vCD to create the catalog item (also creates a transfer folder in the spool area and as result will give a sparse catalog item resource XML).
	//	2. Wait for the links to the transfer folder to appear in the resource representation of the catalog item.
	//	3. Start uploading bits to the transfer folder
	//	4. Wait on the import task to finish on vCD side -> task success = upload complete

	if *cat == (Catalog{}) {
		return UploadTask{}, errors.New("catalog can not be empty or nil")
	}

	ovaFileName, err := validateAndFixFilePath(ovaFileName)
	if err != nil {
		return UploadTask{}, err
	}

	for _, catalogItemName := range getExistingCatalogItems(cat) {
		if catalogItemName == itemName {
			return UploadTask{}, fmt.Errorf("catalog item '%s' already exists. Upload with different name", itemName)
		}
	}

	filesAbsPaths, tmpDir, err := util.Unpack(ovaFileName)
	if err != nil {
		return UploadTask{}, fmt.Errorf("%v. Unpacked files for checking are accessible in: "+tmpDir, err)
	}

	ovfFilePath, err := getOvfPath(filesAbsPaths)
	if err != nil {
		return UploadTask{}, fmt.Errorf("%v. Unpacked files for checking are accessible in: "+tmpDir, err)
	}

	ovfFileDesc, err := getOvf(ovfFilePath)
	if err != nil {
		return UploadTask{}, fmt.Errorf("%v. Unpacked files for checking are accessible in: "+tmpDir, err)
	}

	err = validateOvaContent(filesAbsPaths, &ovfFileDesc, tmpDir)
	if err != nil {
		return UploadTask{}, fmt.Errorf("%v. Unpacked files for checking are accessible in: "+tmpDir, err)
	}

	catalogItemUploadURL, err := findCatalogItemUploadLink(cat, "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml")
	if err != nil {
		return UploadTask{}, err
	}

	vappTemplateUrl, err := createItemForUpload(cat.client, catalogItemUploadURL, itemName, description)
	if err != nil {
		return UploadTask{}, err
	}

	vappTemplate, err := queryVappTemplate(cat.client, vappTemplateUrl, itemName)
	if err != nil {
		return UploadTask{}, err
	}

	ovfUploadHref, err := getUploadLink(vappTemplate.Files)
	if err != nil {
		return UploadTask{}, err
	}

	err = uploadOvfDescription(cat.client, ovfFilePath, ovfUploadHref)
	if err != nil {
		removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
		return UploadTask{}, err
	}

	vappTemplate, err = waitForTempUploadLinks(cat.client, vappTemplateUrl, itemName)
	if err != nil {
		removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
		return UploadTask{}, err
	}

	callBack, uploadProgress := getCallBackFunction()

	uploadError := *new(error)

	//sending upload process to background, this allows no to lock and return task to client
	go uploadFiles(cat.client, vappTemplate, &ovfFileDesc, tmpDir, filesAbsPaths, uploadPieceSize, callBack, &uploadError)

	var task Task
	for _, item := range vappTemplate.Tasks.Task {
		task, err = createTaskForVcdImport(cat.client, item.HREF)
		if err != nil {
			removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
			return UploadTask{}, err
		}
		if task.Task.Status == "error" {
			removeCatalogItemOnError(cat.client, vappTemplateUrl, itemName)
			return UploadTask{}, fmt.Errorf("task did not complete succesfully: %s", task.Task.Description)
		}
	}

	uploadTask := NewUploadTask(&task, uploadProgress, &uploadError)

	util.Logger.Printf("[TRACE] Upload finished and task for vcd import created. \n")

	return *uploadTask, nil
}

// Upload files for vCD created upload links. Different approach then vmdk file are
// chunked (e.g. test.vmdk.000000000, test.vmdk.000000001 or test.vmdk). vmdk files are chunked if
// in description file attribute ChunkSize is not zero.
// params:
// client - client for requests
// vappTemplate - parsed from response vApp template
// ovfFileDesc - parsed from xml part containing ova files definition
// tempPath - path where extracted files are
// filesAbsPaths - array of extracted files
// uploadPieceSize - size of chunks in which the file will be uploaded to the catalog.
// callBack a function with signature //function(bytesUpload, totalSize) to let the caller monitor progress of the upload operation.
// uploadError - error to be ready be task
func uploadFiles(client *Client, vappTemplate *types.VAppTemplate, ovfFileDesc *Envelope, tempPath string, filesAbsPaths []string, uploadPieceSize int64, callBack func(bytesUpload, totalSize int64), uploadError *error) error {
	var uploadedBytes int64
	for _, item := range vappTemplate.Files.File {
		if item.BytesTransferred == 0 {
			number, err := getFileFromDescription(item.Name, ovfFileDesc)
			if err != nil {
				util.Logger.Printf("[Error] Error uploading files: %#v", err)
				*uploadError = err
				return err
			}
			if ovfFileDesc.File[number].ChunkSize != 0 {
				chunkFilePaths := getChunkedFilePaths(tempPath, ovfFileDesc.File[number].HREF, ovfFileDesc.File[number].Size, ovfFileDesc.File[number].ChunkSize)
				details := uploadDetails{
					uploadLink:               item.Link[0].HREF,
					uploadedBytes:            uploadedBytes,
					fileSizeToUpload:         int64(ovfFileDesc.File[number].Size),
					uploadPieceSize:          uploadPieceSize,
					uploadedBytesForCallback: uploadedBytes,
					allFilesSize:             getAllFileSizeSum(ovfFileDesc),
					callBack:                 callBack,
					uploadError:              uploadError,
				}
				tempVar, err := uploadMultiPartFile(client, chunkFilePaths, details)
				if err != nil {
					util.Logger.Printf("[Error] Error uploading files: %#v", err)
					*uploadError = err
					return err
				}
				uploadedBytes += tempVar
			} else {
				details := uploadDetails{
					uploadLink:               item.Link[0].HREF,
					uploadedBytes:            0,
					fileSizeToUpload:         item.Size,
					uploadPieceSize:          uploadPieceSize,
					uploadedBytesForCallback: uploadedBytes,
					allFilesSize:             getAllFileSizeSum(ovfFileDesc),
					callBack:                 callBack,
					uploadError:              uploadError,
				}
				tempVar, err := uploadFile(client, findFilePath(filesAbsPaths, item.Name), details)
				if err != nil {
					util.Logger.Printf("[Error] Error uploading files: %#v", err)
					*uploadError = err
					return err
				}
				uploadedBytes += tempVar
			}
		}
	}

	//remove extracted files with temp dir
	err := os.RemoveAll(tempPath)
	if err != nil {
		util.Logger.Printf("[Error] Error removing temporary files: %#v", err)
		*uploadError = err
		return err
	}

	return nil
}

func getFileFromDescription(fileToFind string, ovfFileDesc *Envelope) (int, error) {
	for fileInArray, item := range ovfFileDesc.File {
		if item.HREF == fileToFind {
			util.Logger.Printf("[TRACE] getFileFromDescription - found matching file: %s in array: %d\n", fileToFind, fileInArray)
			return fileInArray, nil
		}
	}
	return -1, errors.New("file expected from vcd didn't match any description file")
}

func getAllFileSizeSum(ovfFileDesc *Envelope) (sizeSum int64) {
	sizeSum = 0
	for _, item := range ovfFileDesc.File {
		sizeSum += int64(item.Size)
	}
	return
}

// Uploads chunked ova file for vCD created upload link.
// params:
// client - client for requests
// vappTemplate - parsed from response vApp template
// filePaths - all chunked vmdk file paths
// uploadDetails - file upload settings and data
func uploadMultiPartFile(client *Client, filePaths []string, uDetails uploadDetails) (int64, error) {
	util.Logger.Printf("[TRACE] Upload multi part file: %v\n, href: %s, size: %v", filePaths, uDetails.uploadLink, uDetails.fileSizeToUpload)

	var uploadedBytes int64

	for i, filePath := range filePaths {
		util.Logger.Printf("[TRACE] Uploading file: %v\n", i+1)
		uDetails.uploadedBytesForCallback += uploadedBytes // previous files uploaded size plus current upload size
		uDetails.uploadedBytes = uploadedBytes
		tempVar, err := uploadFile(client, filePath, uDetails)
		if err != nil {
			return uploadedBytes, err
		}
		uploadedBytes += tempVar
	}
	return uploadedBytes, nil
}

// Function waits until vCD provides temporary file upload links.
func waitForTempUploadLinks(client *Client, vappTemplateUrl *url.URL, newItemName string) (*types.VAppTemplate, error) {
	var vAppTemplate *types.VAppTemplate
	var err error
	for {
		util.Logger.Printf("[TRACE] Sleep... for 5 seconds.\n")
		time.Sleep(time.Second * 5)
		vAppTemplate, err = queryVappTemplate(client, vappTemplateUrl, newItemName)
		if err != nil {
			return nil, err
		}
		if vAppTemplate.Files != nil && len(vAppTemplate.Files.File) > 1 {
			util.Logger.Printf("[TRACE] upload link prepared.\n")
			break
		}
	}
	return vAppTemplate, nil
}

func queryVappTemplate(client *Client, vappTemplateUrl *url.URL, newItemName string) (*types.VAppTemplate, error) {
	util.Logger.Printf("[TRACE] Querying vapp template: %s\n", vappTemplateUrl)
	request := client.NewRequest(map[string]string{}, "GET", *vappTemplateUrl, nil)
	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return nil, err
	}

	vappTemplateParsed := &types.VAppTemplate{}
	if err = decodeBody(response, vappTemplateParsed); err != nil {
		return nil, err
	}

	defer response.Body.Close()

	for _, task := range vappTemplateParsed.Tasks.Task {
		if "error" == task.Status && newItemName == task.Owner.Name {
			util.Logger.Printf("[Error] %#v", task.Error)
			return vappTemplateParsed, fmt.Errorf("Error in vcd returned error code: %d, error: %s and message: %s ", task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
		}
	}

	return vappTemplateParsed, nil
}

// Uploads ovf description file from unarchived provided ova file. As a result vCD will generate temporary upload links which has to be queried later.
// Function will return parsed part for upload files from description xml.
func uploadOvfDescription(client *Client, ovfFile string, ovfUploadUrl *url.URL) error {
	util.Logger.Printf("[TRACE] Uploding ovf description with file: %s and url: %s\n", ovfFile, ovfUploadUrl)
	openedFile, err := os.Open(ovfFile)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	ovfReader := io.TeeReader(openedFile, &buf)

	request := client.NewRequest(map[string]string{}, "PUT", *ovfUploadUrl, ovfReader)
	request.Header.Add("Content-Type", "text/xml")

	_, err = checkResp(client.Http.Do(request))
	if err != nil {
		return err
	}

	err = openedFile.Close()
	if err != nil {
		util.Logger.Printf("[Error] Error closing file: %#v", err)
		return err
	}

	return nil
}

func parseOvfFileDesc(file *os.File, ovfFileDesc *Envelope) error {
	ovfXml, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(ovfXml, &ovfFileDesc)
	if err != nil {
		return err
	}
	return nil
}

func findCatalogItemUploadLink(catalog *Catalog, applicationType string) (*url.URL, error) {
	for _, item := range catalog.Catalog.Link {
		if item.Type == applicationType && item.Rel == "add" {
			util.Logger.Printf("[TRACE] Found Catalong link for upload: %s\n", item.HREF)

			uploadURL, err := url.ParseRequestURI(item.HREF)
			if err != nil {
				return nil, err
			}

			util.Logger.Printf("[TRACE] findCatalogItemUploadLink - catalog item upload url found: %s \n", uploadURL)
			return uploadURL, nil
		}
	}
	return nil, errors.New("catalog upload URL not found")
}

func getExistingCatalogItems(catalog *Catalog) (catalogItemNames []string) {
	for _, catalogItems := range catalog.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			catalogItemNames = append(catalogItemNames, catalogItem.Name)
		}
	}
	return
}

func findFilePath(filesAbsPaths []string, fileName string) string {
	for _, item := range filesAbsPaths {
		_, file := filepath.Split(item)
		if file == fileName {
			return item
		}
	}
	return ""
}

// Initiates creation of item and returns ovf upload url for created item.
func createItemForUpload(client *Client, createHREF *url.URL, catalogItemName string, itemDescription string) (*url.URL, error) {
	util.Logger.Printf("[TRACE] createItemForUpload: %s, item name: %v, description: %v \n", createHREF, catalogItemName, itemDescription)
	reqBody := bytes.NewBufferString(
		"<UploadVAppTemplateParams xmlns=\"http://www.vmware.com/vcloud/v1.5\" name=\"" + catalogItemName + "\" >" +
			"<Description>" + itemDescription + "</Description>" +
			"</UploadVAppTemplateParams>")

	request := client.NewRequest(map[string]string{}, "POST", *createHREF, reqBody)
	request.Header.Add("Content-Type", "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml")

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	catalogItemParsed := &types.CatalogItem{}
	if err = decodeBody(response, catalogItemParsed); err != nil {
		return nil, err
	}

	util.Logger.Printf("[TRACE] Catalog item parsed: %#v\n", catalogItemParsed)

	ovfUploadUrl, err := url.ParseRequestURI(catalogItemParsed.Entity.HREF)
	if err != nil {
		return nil, err
	}

	return ovfUploadUrl, nil
}

// Helper method to get path to multi-part files.
//For example a file called test.vmdk with total_file_size = 100 bytes and part_size = 40 bytes, implies the file is made of *3* part files.
//		- test.vmdk.000000000 = 40 bytes
//		- test.vmdk.000000001 = 40 bytes
//		- test.vmdk.000000002 = 20 bytes
//Say base_dir = /dummy_path/, and base_file_name = test.vmdk then
//the output of this function will be [/dummy_path/test.vmdk.000000000,
// /dummy_path/test.vmdk.000000001, /dummy_path/test.vmdk.000000002]
func getChunkedFilePaths(baseDir, baseFileName string, totalFileSize, partSize int) []string {
	var filePaths []string
	numbParts := math.Ceil(float64(totalFileSize) / float64(partSize))
	for i := 0; i < int(numbParts); i++ {
		temp := "000000000" + strconv.Itoa(i)
		postfix := temp[len(temp)-9:]
		filePath := path.Join(baseDir, baseFileName+"."+postfix)
		filePaths = append(filePaths, filePath)
	}

	util.Logger.Printf("[TRACE] Chunked files file paths: %s \n", filePaths)
	return filePaths
}

func getOvfPath(filesAbsPaths []string) (string, error) {
	for _, filePath := range filesAbsPaths {
		if filepath.Ext(filePath) == ".ovf" {
			return filePath, nil
		}
	}
	return "", errors.New("ova is not correct - missing ovf file")
}

func getOvf(ovfFilePath string) (Envelope, error) {
	openedFile, err := os.Open(ovfFilePath)
	if err != nil {
		return Envelope{}, err
	}

	var ovfFileDesc Envelope
	err = parseOvfFileDesc(openedFile, &ovfFileDesc)
	if err != nil {
		return Envelope{}, err
	}

	err = openedFile.Close()
	if err != nil {
		util.Logger.Printf("[Error] Error closing file: %#v", err)
		return Envelope{}, err
	}

	return ovfFileDesc, nil
}

func validateOvaContent(filesAbsPaths []string, ovfFileDesc *Envelope, tempPath string) error {
	for _, fileDescription := range ovfFileDesc.File {
		if fileDescription.ChunkSize == 0 {
			err := checkIfFileMatchesDescription(filesAbsPaths, fileDescription)
			if err != nil {
				return err
			}
			// check chunked ova content
		} else {
			chunkFilePaths := getChunkedFilePaths(tempPath, fileDescription.HREF, fileDescription.Size, fileDescription.ChunkSize)
			for part, chunkedFilePath := range chunkFilePaths {
				_, fileName := filepath.Split(chunkedFilePath)
				chunkedFileSize := fileDescription.Size - part*fileDescription.ChunkSize
				if chunkedFileSize > fileDescription.ChunkSize {
					chunkedFileSize = fileDescription.ChunkSize
				}
				chunkedFileDescription := struct {
					HREF      string `xml:"href,attr"`
					ID        string `xml:"id,attr"`
					Size      int    `xml:"size,attr"`
					ChunkSize int    `xml:"chunkSize,attr"`
				}{fileName, "", chunkedFileSize, fileDescription.ChunkSize}
				err := checkIfFileMatchesDescription(filesAbsPaths, chunkedFileDescription)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkIfFileMatchesDescription(filesAbsPaths []string, fileDescription struct {
	HREF      string `xml:"href,attr"`
	ID        string `xml:"id,attr"`
	Size      int    `xml:"size,attr"`
	ChunkSize int    `xml:"chunkSize,attr"`
}) error {
	filePath := findFilePath(filesAbsPaths, fileDescription.HREF)
	if filePath == "" {
		return fmt.Errorf("file '%s' described in ovf was not found in ova", fileDescription.HREF)
	}
	if fileInfo, err := os.Stat(filePath); err == nil {
		if fileInfo.Size() != int64(fileDescription.Size) {
			return fmt.Errorf("file size didn't match described in ovf: %s", filePath)
		}
	} else {
		return err
	}
	return nil
}

func removeCatalogItemOnError(client *Client, vappTemplateLink *url.URL, itemName string) {
	if vappTemplateLink != nil {
		util.Logger.Printf("[TRACE] Deleting Catalog item %v", vappTemplateLink)

		// wait for task, cancel it and catalog item will be removed.
		var vAppTemplate *types.VAppTemplate
		var err error
		for {
			util.Logger.Printf("[TRACE] Sleep... for 5 seconds.\n")
			time.Sleep(time.Second * 5)
			vAppTemplate, err = queryVappTemplate(client, vappTemplateLink, itemName)
			if err != nil {
				util.Logger.Printf("[Error] Error deleting Catalog item %v: %s", vappTemplateLink, err)
			}
			if len(vAppTemplate.Tasks.Task) > 0 {
				util.Logger.Printf("[TRACE] Task found. Will try to cancel.\n")
				break
			}
		}

		for _, taskItem := range vAppTemplate.Tasks.Task {
			if itemName == taskItem.Owner.Name {
				task := NewTask(client)
				task.Task = taskItem
				err = task.CancelTask()
				if err != nil {
					util.Logger.Printf("[ERROR] Error canceling task for catalog item upload %#v", err)
				}
			}
		}
	} else {
		util.Logger.Printf("[Error] Failed to delete catalog item created with error: %v", vappTemplateLink)
	}
}

func (cat *Catalog) UploadMediaImage(mediaName, mediaDescription, filePath string, uploadPieceSize int64) (UploadTask, error) {

	if *cat == (Catalog{}) {
		return UploadTask{}, errors.New("catalog can not be empty or nil")
	}

	mediaFilePath, err := validateAndFixFilePath(filePath)
	if err != nil {
		return UploadTask{}, err
	}

	isISOGood, err := verifyIso(mediaFilePath)
	if err != nil || !isISOGood {
		return UploadTask{}, fmt.Errorf("[ERROR] File %s isn't correct iso file: %#v", mediaFilePath, err)
	}

	file, e := os.Stat(mediaFilePath)
	if e != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue finding file: %#v", e)
	}
	fileSize := file.Size()

	for _, catalogItemName := range getExistingCatalogItems(cat) {
		if catalogItemName == mediaName {
			return UploadTask{}, fmt.Errorf("media item '%s' already exists. Upload with different name", mediaName)
		}
	}

	catalogItemUploadURL, err := findCatalogItemUploadLink(cat, "application/vnd.vmware.vcloud.media+xml")
	if err != nil {
		return UploadTask{}, err
	}

	mediaItem, err := createMedia(cat.client, catalogItemUploadURL.String(), mediaName, mediaDescription, fileSize)
	if err != nil {
		return UploadTask{}, fmt.Errorf("[ERROR] Issue creating media: %#v", err)
	}

	createdMedia, err := queryMedia(cat.client, mediaItem.Entity.HREF, mediaName)
	if err != nil {
		return UploadTask{}, err
	}

	return executeUpload(cat.client, createdMedia, mediaFilePath, mediaName, fileSize, uploadPieceSize)
}
