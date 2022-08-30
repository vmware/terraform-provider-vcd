package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func lookupvAppTemplateforVm(d *schema.ResourceData, org *govcd.Org, vdc *govcd.Vdc) (govcd.VAppTemplate, error) {
	catalogName := d.Get("catalog_name").(string)
	templateName := d.Get("template_name").(string)

	catalog, err := org.GetCatalogByName(catalogName, false)
	if err != nil {
		return govcd.VAppTemplate{}, fmt.Errorf("error finding catalog %s: %s", catalogName, err)
	}

	var vappTemplate govcd.VAppTemplate
	if vmNameInTemplate, ok := d.GetOk("vm_name_in_template"); ok {
		vmInTemplateRecord, err := vdc.QueryVappVmTemplate(catalogName, templateName, vmNameInTemplate.(string))
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error quering VM template %s: %s", vmNameInTemplate, err)
		}
		util.Logger.Printf("[VM create] vmInTemplateRecord %# v", pretty.Formatter(vmInTemplateRecord))
		returnedVappTemplate, err := catalog.GetVappTemplateByHref(vmInTemplateRecord.HREF)
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error quering VM template %s: %s", vmNameInTemplate, err)
		}
		util.Logger.Printf("[VM create] returned VappTemplate %#v", pretty.Formatter(returnedVappTemplate))
		vappTemplate = *returnedVappTemplate
	} else {
		catalogItem, err := catalog.GetCatalogItemByName(templateName, false)
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("error finding catalog item %s: %s", templateName, err)
		}
		vappTemplate, err = catalogItem.GetVAppTemplate()
		if err != nil {
			return govcd.VAppTemplate{}, fmt.Errorf("[VM create] error finding VAppTemplate %s: %s", templateName, err)
		}

	}

	return vappTemplate, nil
}

func lookupStorageProfile(d *schema.ResourceData, vdc *govcd.Vdc) (*types.Reference, error) {
	// If no storage profile lookup was requested - bail out early and return nil reference
	storageProfileName := d.Get("storage_profile").(string)
	if storageProfileName == "" {
		return nil, nil
	}

	storageProfile, err := vdc.FindStorageProfileReference(storageProfileName)
	if err != nil {
		return nil, fmt.Errorf("[vm creation] error retrieving storage profile %s : %s", storageProfileName, err)
	}

	return &storageProfile, nil

}

func lookupComputePolicy(d *schema.ResourceData, vcdClient *VCDClient) (*types.VdcComputePolicy, *types.ComputePolicy, error) {
	var sizingPolicy *types.VdcComputePolicy
	var vmComputePolicy *types.ComputePolicy
	if value, ok := d.GetOk("sizing_policy_id"); ok {
		vdcComputePolicy, err := vcdClient.Client.GetVdcComputePolicyById(value.(string))
		if err != nil {
			return nil, nil, fmt.Errorf("error getting sizing policy %s: %s", value.(string), err)
		}
		sizingPolicy = vdcComputePolicy.VdcComputePolicy
		if vdcComputePolicy.Href == "" {
			return nil, nil, fmt.Errorf("empty sizing policy HREF detected")
		}
		vmComputePolicy = &types.ComputePolicy{
			VmSizingPolicy: &types.Reference{HREF: vdcComputePolicy.Href},
		}
		util.Logger.Printf("[VM create] sizingPolicy (%s) %# v", vdcComputePolicy.Href, pretty.Formatter(sizingPolicy))
	}

	return sizingPolicy, vmComputePolicy, nil
}
