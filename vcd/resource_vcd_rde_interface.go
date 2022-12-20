package vcd

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceVcdRdeInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: nil,
		ReadContext:   nil,
		UpdateContext: nil,
		DeleteContext: nil,
		Importer: &schema.ResourceImporter{
			StateContext: nil,
		},
		Schema: map[string]*schema.Schema{},
	}
}
