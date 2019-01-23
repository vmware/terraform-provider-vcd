/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/types/v56"
)

type VAppTemplate struct {
	VAppTemplate *types.VAppTemplate
	client       *Client
}

func NewVAppTemplate(cli *Client) *VAppTemplate {
	return &VAppTemplate{
		VAppTemplate: new(types.VAppTemplate),
		client:       cli,
	}
}

func (vdc *Vdc) InstantiateVAppTemplate(template *types.InstantiateVAppTemplateParams) error {
	output, err := xml.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("Error finding VAppTemplate: %#v", err)
	}
	requestData := bytes.NewBufferString(xml.Header + string(output))

	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/instantiateVAppTemplate"

	req := vdc.client.NewRequest(map[string]string{}, "POST", *vdcHref, requestData)
	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml")

	resp, err := checkResp(vdc.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error instantiating a new template: %s", err)
	}

	vapptemplate := NewVAppTemplate(vdc.client)
	if err = decodeBody(resp, vapptemplate.VAppTemplate); err != nil {
		return fmt.Errorf("error decoding orgvdcnetwork response: %s", err)
	}
	task := NewTask(vdc.client)
	for _, taskItem := range vapptemplate.VAppTemplate.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Error performing task: %#v", err)
		}
	}
	return nil
}
