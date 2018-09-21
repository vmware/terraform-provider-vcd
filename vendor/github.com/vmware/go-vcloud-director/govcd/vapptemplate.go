/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"net/url"
)

type VAppTemplate struct {
	VAppTemplate *types.VAppTemplate
	c            *Client
}

func NewVAppTemplate(c *Client) *VAppTemplate {
	return &VAppTemplate{
		VAppTemplate: new(types.VAppTemplate),
		c:            c,
	}
}

func (v *Vdc) InstantiateVAppTemplate(template *types.InstantiateVAppTemplateParams) error {
	output, err := xml.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("Error finding VAppTemplate: %#v", err)
	}
	requestData := bytes.NewBufferString(xml.Header + string(output))

	vdcHref, err := url.ParseRequestURI(v.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/instantiateVAppTemplate"

	req := v.c.NewRequest(map[string]string{}, "POST", *vdcHref, requestData)
	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.instantiateVAppTemplateParams+xml")

	resp, err := checkResp(v.c.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error instantiating a new template: %s", err)
	}

	vapptemplate := NewVAppTemplate(v.c)
	if err = decodeBody(resp, vapptemplate.VAppTemplate); err != nil {
		return fmt.Errorf("error decoding orgvdcnetwork response: %s", err)
	}
	task := NewTask(v.c)
	for _, t := range vapptemplate.VAppTemplate.Tasks.Task {
		task.Task = t
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("Error performing task: %#v", err)
		}
	}
	return nil
}
