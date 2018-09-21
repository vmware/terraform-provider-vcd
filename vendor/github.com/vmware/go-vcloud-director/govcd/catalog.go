/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/vmware/go-vcloud-director/types/v56"
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

type Catalog struct {
	Catalog *types.Catalog
	c       *Client
}

func NewCatalog(c *Client) *Catalog {
	return &Catalog{
		Catalog: new(types.Catalog),
		c:       c,
	}
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

func (c *Catalog) FindCatalogItem(catalogitem string) (CatalogItem, error) {

	for _, cis := range c.Catalog.CatalogItems {
		for _, ci := range cis.CatalogItem {
			if ci.Name == catalogitem && ci.Type == "application/vnd.vmware.vcloud.catalogItem+xml" {
				u, err := url.ParseRequestURI(ci.HREF)

				if err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				req := c.c.NewRequest(map[string]string{}, "GET", *u, nil)

				resp, err := checkResp(c.c.Http.Do(req))
				if err != nil {
					return CatalogItem{}, fmt.Errorf("error retreiving catalog: %s", err)
				}

				cat := NewCatalogItem(c.c)

				if err = decodeBody(resp, cat.CatalogItem); err != nil {
					return CatalogItem{}, fmt.Errorf("error decoding catalog response: %s", err)
				}

				// The request was successful
				return *cat, nil
			}
		}
	}

	return CatalogItem{}, fmt.Errorf("can't find catalog item: %s", catalogitem)
}

// uploads an ova file to a catalog. This method only uploads bits to vCD spool area.
// Returns errors if any occur during upload from vCD or upload process.
func (c *Catalog) UploadOvf(ovaFileName, itemName, description string, chunkSize int) (Task, error) {

	//	On a very high level the flow is as follows
	//	1. Makes a POST call to vCD to create the catalog item (also creates a transfer folder in the spool area and as result will give a sparse catalog item resource XML).
	//	2. Wait for the links to the transfer folder to appear in the resource representation of the catalog item.
	//	3. Start uploading bits to the transfer folder
	//	4. Wait on the import task to finish on vCD side -> task success = upload complete

	catalogItemUploadURL, err := findCatalogItemUploadLink(c)
	if err != nil {
		return Task{}, err
	}

	vappTemplateUrl, err := createItemForUpload(c.c, catalogItemUploadURL, itemName, description)
	if err != nil {
		return Task{}, err
	}

	vappTemplate, err := queryVappTemplate(c.c, vappTemplateUrl)
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
			ovfFileDesc, err = uploadOvfDescription(c.c, filePath, ovfUploadHref)
			tempPath, _ = filepath.Split(filePath)
			if err != nil {
				return Task{}, err
			}
			break
		}
	}

	vappTemplate, err = waitForTempUploadLinks(c.c, vappTemplateUrl)

	err = uploadFiles(c.c, vappTemplate, &ovfFileDesc, tempPath, filesAbsPaths)
	if err != nil {
		return Task{}, err
	}

	var task Task
	for _, item := range vappTemplate.Tasks.Task {
		task, err = createTaskForVcdImport(c.c, item.HREF)
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
