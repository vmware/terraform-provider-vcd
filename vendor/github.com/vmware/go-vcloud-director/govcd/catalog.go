/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type CatalogOperations interface {
	FindCatalogItem(catalogitem string) (CatalogItem, error)
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
func (adminCatalog *AdminCatalog) Delete(force, recursive bool) error {
	adminCatalogHREF := adminCatalog.client.VCDHREF
	adminCatalogHREF.Path += "/admin/catalog/" + adminCatalog.AdminCatalog.ID[19:]

	req := adminCatalog.client.NewRequest(map[string]string{
		"force":     strconv.FormatBool(force),
		"recursive": strconv.FormatBool(recursive),
	}, "DELETE", adminCatalogHREF, nil)

	_, err := checkResp(adminCatalog.client.Http.Do(req))

	if err != nil {
		return fmt.Errorf("error deleting Catalog %s: %s", adminCatalog.AdminCatalog.ID, err)
	}

	return nil
}

//   Updates the Catalog definition from current Catalog struct contents.
//   Any differences that may be legally applied will be updated.
//   Returns an error if the call to vCD fails. Update automatically performs
//   a refresh with the admin catalog it gets back from the rest api
//   Link to API call: https://code.vmware.com/apis/220/vcloud#/doc/doc/operations/PUT-Catalog.html
func (adminCatalog *AdminCatalog) Update() error {
	vcomp := &types.AdminCatalog{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        adminCatalog.AdminCatalog.Name,
		Description: adminCatalog.AdminCatalog.Description,
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

// If catalogitem is a valid CatalogItem and the call succeds,
// then the function returns a CatalogItem. If the item does not
// exist, then it returns an empty CatalogItem. If The call fails
// at any point, it returns an error.
func (cat *Catalog) FindCatalogItem(catalogitem string) (CatalogItem, error) {
	for _, catalogItems := range cat.Catalog.CatalogItems {
		for _, catalogItem := range catalogItems.CatalogItem {
			if catalogItem.Name == catalogitem && catalogItem.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				catalogItemHREF, err := url.ParseRequestURI(catalogItem.HREF)

				if err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				req := cat.client.NewRequest(map[string]string{}, "GET", *catalogItemHREF, nil)

				resp, err := checkResp(cat.client.Http.Do(req))
				if err != nil {
					return CatalogItem{}, fmt.Errorf("error retreiving catalog: %s", err)
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

// uploads an ova file to a catalog. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process.
func (cat *Catalog) UploadOvf(ovaFileName, itemName, description string, chunkSize int) (Task, error) {

	//	On a very high level the flow is as follows
	//	1. Makes a POST call to vCD to create the catalog item (also creates a transfer folder in the spool area and as result will give a sparse catalog item resource XML).
	//	2. Wait for the links to the transfer folder to appear in the resource representation of the catalog item.
	//	3. Start uploading bits to the transfer folder
	//	4. Wait on the import task to finish on vCD side -> task success = upload complete

	catalogItemUploadURL, err := findCatalogItemUploadLink(cat)
	if err != nil {
		return Task{}, err
	}

	vappTemplateUrl, err := createItemForUpload(cat.client, catalogItemUploadURL, itemName, description)
	if err != nil {
		return Task{}, err
	}

	vappTemplate, err := queryVappTemplate(cat.client, vappTemplateUrl)
	if err != nil {
		return Task{}, err
	}

	ovfUploadHref, err := getOvfUploadLink(vappTemplate)
	if err != nil {
		return Task{}, err
	}

	filesAbsPaths, err := util.Unpack(ovaFileName)
	if err != nil {
		return Task{}, err
	}

	var ovfFileDesc Envelope
	var tempPath string

	for _, filePath := range filesAbsPaths {
		if filepath.Ext(filePath) == ".ovf" {
			ovfFileDesc, err = uploadOvfDescription(cat.client, filePath, ovfUploadHref)
			tempPath, _ = filepath.Split(filePath)
			if err != nil {
				return Task{}, err
			}
			break
		}
	}

	vappTemplate, err = waitForTempUploadLinks(cat.client, vappTemplateUrl)

	err = uploadFiles(cat.client, vappTemplate, &ovfFileDesc, tempPath, filesAbsPaths)
	if err != nil {
		return Task{}, err
	}

	var task Task
	for _, item := range vappTemplate.Tasks.Task {
		task, err = createTaskForVcdImport(cat.client, item.HREF)
	}

	log.Printf("[TRACE] Upload finished and task for vcd import created. \n")
	return task, nil
}

func uploadFiles(client *Client, vappTemplate *types.VAppTemplate, ovfFileDesc *Envelope, tempPath string, filesAbsPaths []string) error {
	for _, item := range vappTemplate.Files.File {
		if item.BytesTransferred == 0 {
			if ovfFileDesc.File[0].ChunkSize != 0 {
				chunkFilePaths := getChunkedFilePaths(tempPath, ovfFileDesc.File[0].HREF, ovfFileDesc.File[0].Size, ovfFileDesc.File[0].ChunkSize)
				err := uploadMultiPartFile(client, chunkFilePaths, item.Link[0].HREF, int64(ovfFileDesc.File[0].Size))
				if err != nil {
					return err
				}
			} else {
				_, err := uploadFile(client, item.Link[0].HREF, findFilePath(filesAbsPaths, item.Name), 0, item.Size)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func uploadMultiPartFile(client *Client, filePaths []string, uploadHREF string, totalBytesToUpload int64) error {
	log.Printf("[TRACE] Upload multi part file: %v\n, href: %s, size: %v", filePaths, uploadHREF, totalBytesToUpload)

	var uploadedBytes int64

	for i, filePath := range filePaths {
		log.Printf("[TRACE] Uploading file: %v\n", i+1)
		tempVar, err := uploadFile(client, uploadHREF, filePath, uploadedBytes, totalBytesToUpload)
		if err != nil {
			return err
		}
		uploadedBytes += tempVar
	}
	return nil
}

// Function waits until vCD provides temporary file upload links.
func waitForTempUploadLinks(client *Client, vappTemplateUrl *url.URL) (*types.VAppTemplate, error) {
	var vAppTemplate *types.VAppTemplate
	var err error
	for {
		log.Printf("[TRACE] Sleep... for 5 seconds.\n")
		time.Sleep(time.Second * 5)
		vAppTemplate, err = queryVappTemplate(client, vappTemplateUrl)
		if err != nil {
			return nil, err
		}
		if len(vAppTemplate.Files.File) > 1 {
			log.Printf("[TRACE] upload link prepared.\n")
			break
		}
	}
	return vAppTemplate, nil
}

func createTaskForVcdImport(client *Client, taskHREF string) (Task, error) {
	log.Printf("[TRACE] Create task for vcd with HREF: %s\n", taskHREF)

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

func getOvfUploadLink(vappTemplate *types.VAppTemplate) (*url.URL, error) {
	log.Printf("[TRACE] Parsing ofv upload link: %#v\n", vappTemplate)

	ovfUploadHref, err := url.ParseRequestURI(vappTemplate.Files.File[0].Link[0].HREF)
	if err != nil {
		return nil, err
	}

	return ovfUploadHref, nil
}

func queryVappTemplate(client *Client, vappTemplateUrl *url.URL) (*types.VAppTemplate, error) {
	log.Printf("[TRACE] Qeurying vapp template: %s\n", vappTemplateUrl)
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

	log.Printf("[TRACE] Response: %v\n", response)
	log.Printf("[TRACE] Response body: %v\n", vappTemplateParsed)
	return vappTemplateParsed, nil
}

// Uploads ovf description file from unarchived provided ova file. As result vCD will generate temporary upload links which has to be queried later.
// Function will return parsed part for upload files from description xml.
func uploadOvfDescription(client *Client, ovfFile string, ovfUploadUrl *url.URL) (Envelope, error) {
	log.Printf("[TRACE] Uploding ovf description with file: %s and url: %s\n", ovfFile, ovfUploadUrl)
	openedFile, err := os.Open(ovfFile)
	if err != nil {
		return Envelope{}, err
	}

	var buf bytes.Buffer
	ovfReader := io.TeeReader(openedFile, &buf)

	request := client.NewRequest(map[string]string{}, "PUT", *ovfUploadUrl, ovfReader)
	request.Header.Add("Content-Type", "text/xml")

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return Envelope{}, err
	}

	var ovfFileDesc Envelope
	ovfXml, err := ioutil.ReadAll(&buf)
	if err != nil {
		return Envelope{}, err
	}

	err = xml.Unmarshal(ovfXml, &ovfFileDesc)
	if err != nil {
		return Envelope{}, err
	}

	openedFile.Close()

	body, err := ioutil.ReadAll(response.Body)
	log.Printf("[TRACE] Response: %#v\n", response)
	log.Printf("[TRACE] Response body: %s\n", string(body[:]))
	log.Printf("[TRACE] Ovf file description file: %#v\n", ovfFileDesc)

	response.Body.Close()

	return ovfFileDesc, nil
}

func findCatalogItemUploadLink(catalog *Catalog) (*url.URL, error) {
	for _, item := range catalog.Catalog.Link {
		if item.Type == "application/vnd.vmware.vcloud.uploadVAppTemplateParams+xml" && item.Rel == "add" {
			log.Printf("[TRACE] Found Catalong link for uplaod: %s\n", item.HREF)

			uploadURL, err := url.ParseRequestURI(item.HREF)
			if err != nil {
				return nil, err
			}

			return uploadURL, nil
		}
	}
	return nil, errors.New("catalog upload url isn't found")
}

func uploadFile(client *Client, uploadLink, filePath string, offset, fileSizeToUpload int64) (int64, error) {
	log.Printf("[TRACE] Starting uploading: %s, offset: %v, fileze: %v, toLink: %s \n", filePath, offset, fileSizeToUpload, uploadLink)

	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}

	defer file.Close()

	request, err := newFileUploadRequest(uploadLink, file, offset, fileInfo.Size(), fileSizeToUpload)
	if err != nil {
		return 0, err
	}

	response, err := checkResp(client.Http.Do(request))
	if err != nil {
		return 0, fmt.Errorf("File "+filePath+" upload failed. Err: %s \n", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	log.Printf("[TRACE] Response: %#v\n", response)
	log.Printf("[TRACE] Response body: %s\n", string(body[:]))

	return fileInfo.Size(), nil
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

	log.Printf("[TRACE] Response: %#v \n", response)
	log.Printf("[TRACE] Catalog item parsed: %#v\n", catalogItemParsed)

	ovfUploadUrl, err := url.ParseRequestURI(catalogItemParsed.Entity.HREF)
	if err != nil {
		return nil, err
	}

	return ovfUploadUrl, nil
}

// Create Request with right headers and range settings. Support multi part file upload.
func newFileUploadRequest(requestUrl string, file io.Reader, offset, fileSize, fileSizeToUpload int64) (*http.Request, error) {
	log.Printf("[TRACE] Creating file upload request: %s, %v, %v, %v \n", requestUrl, offset, fileSize, fileSizeToUpload)

	uploadReq, err := http.NewRequest("PUT", requestUrl, file)
	if err != nil {
		return nil, err
	}

	uploadReq.ContentLength = int64(fileSize)
	uploadReq.Header.Set("Content-Length", strconv.FormatInt(uploadReq.ContentLength, 10))

	rangeExpression := "bytes " + strconv.FormatInt(int64(offset), 10) + "-" + strconv.FormatInt(int64(offset+fileSize-1), 10) + "/" + strconv.FormatInt(int64(fileSizeToUpload), 10)
	uploadReq.Header.Set("Content-Range", rangeExpression)

	for key, value := range uploadReq.Header {
		log.Printf("[TRACE] Header: %s :%s \n", key, value)
	}

	return uploadReq, nil
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

	log.Printf("[TRACE] Chunked files file paths: %s \n", filePaths)
	return filePaths
}
