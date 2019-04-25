/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
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
	vdcHref, err := url.ParseRequestURI(vdc.Vdc.HREF)
	if err != nil {
		return fmt.Errorf("error getting vdc href: %v", err)
	}
	vdcHref.Path += "/action/instantiateVAppTemplate"

	vapptemplate := NewVAppTemplate(vdc.client)

	_, err = vdc.client.ExecuteRequest(vdcHref.String(), http.MethodPut,
		types.MimeInstantiateVappTemplateParams, "error instantiating a new template: %s", template, vapptemplate)
	if err != nil {
		return err
	}

	task := NewTask(vdc.client)
	for _, taskItem := range vapptemplate.VAppTemplate.Tasks.Task {
		task.Task = taskItem
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error performing task: %#v", err)
		}
	}
	return nil
}
