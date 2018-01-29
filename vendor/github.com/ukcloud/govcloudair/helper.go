package govcloudair

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"

	types "github.com/ukcloud/govcloudair/types/v56"
)

func GetVersionHeader(version types.ApiVersionType) (key, value string) {
	return "Accept", fmt.Sprintf("application/*+xml;version=%s", version)
}

func ExecuteRequest(payload, path, type_, contentType string, client *Client) (Task, error) {
	s, _ := url.ParseRequestURI(path)

	log.Printf("[TRACE] URL: %s", path)
	log.Printf("[TRACE] Type: %s", type_)
	log.Printf("[TRACE] ContentType: %s", contentType)

	var req *http.Request
	switch type_ {
	case "POST":
		log.Printf("[TRACE] XML: \n %s", payload)

		b := bytes.NewBufferString(xml.Header + payload)
		req = client.NewRequest(map[string]string{}, type_, *s, b)

	default:
		req = client.NewRequest(map[string]string{}, type_, *s, nil)

	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	resp, err := checkResp(client.Http.Do(req))
	if err != nil {
		return Task{}, err
	}

	task := NewTask(client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	// The request was successful
	return *task, nil

}
