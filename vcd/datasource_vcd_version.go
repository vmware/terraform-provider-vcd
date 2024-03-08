package vcd

import (
	"context"
	"fmt"
	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVersionRead,
		Schema: map[string]*schema.Schema{
			"condition": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "A condition to check against the VCD version",
				RequiredWith: []string{"fail_if_not_match"},
			},
			"fail_if_not_match": {
				Type:         schema.TypeBool,
				Optional:     true,
				Description:  "This data source fails if the VCD doesn't match the version constraint set in 'condition'",
				RequiredWith: []string{"condition"},
			},
			"matches_condition": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether VCD matches the condition or not",
			},
			"vcd_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VCD version",
			},
			"api_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VCD API version",
			},
		},
	}
}

func datasourceVcdVersionRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdVersion, err := vcdClient.VCDClient.Client.GetVcdShortVersion()
	if err != nil {
		return diag.Errorf("could not get VCD version: %s", err)
	}
	apiVersion, err := vcdClient.VCDClient.Client.MaxSupportedVersion()
	if err != nil {
		return diag.Errorf("could not get VCD API version: %s", err)
	}

	dSet(d, "vcd_version", vcdVersion)
	dSet(d, "api_version", apiVersion)

	if condition, ok := d.GetOk("condition"); ok {
		checkVer, err := semver.NewVersion(vcdVersion)
		if err != nil {
			return diag.Errorf("unable to parse version '%s': %s", vcdVersion, err)
		}
		constraints, err := semver.NewConstraint(condition.(string))
		if err != nil {
			return diag.Errorf("unable to parse given version constraint '%s' : %s", condition, err)
		}
		matchesCondition := constraints.Check(checkVer)
		dSet(d, "matches_condition", matchesCondition)
		if !matchesCondition && d.Get("fail_if_not_match").(bool) {
			return diag.Errorf("the VCD version '%s' doesn't match the version constraint '%s'", vcdVersion, condition)
		}
	}

	// The ID is artificial, and we try to identify each data source instance unequivocally through its parameters.
	d.SetId(fmt.Sprintf("vcd_version='%s',condition='%s',fail_if_not_match='%t'", vcdVersion, d.Get("condition"), d.Get("fail_if_not_match")))
	return nil
}
