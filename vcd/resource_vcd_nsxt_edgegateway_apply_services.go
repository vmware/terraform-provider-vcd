package vcd

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// resourceVcdNsxtEdgeGatewayIpSpaceServices
func resourceVcdNsxtEdgeGatewayIpSpaceServices() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtEdgeGatewayIpSpaceServicesCreate,
		ReadContext:   resourceVcdNsxtEdgeGatewayIpSpaceServicesRead,
		UpdateContext: resourceVcdNsxtEdgeGatewayIpSpaceServicesUpdate,
		DeleteContext: noOp,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID for Firewall and NAT rule autoconfiguration",
			},

			"trigger_on_create": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Will trigger only once on creation",
				// ConflictsWith: []string{"trigger_on_every_apply"},
				ExactlyOneOf: []string{"trigger_on_create", "trigger_on_every_apply"},
			},
			"trigger_on_every_apply": {
				ValidateFunc: noopValueWarningValidator(true,
					"Using 'true' value for field 'trigger_on_every_apply' will cause plan change "+
						"every time and might cause an error if rules already exist. Field "+
						"'ignore_existing_rule_error' can be used to ignore such error."),
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				// Default:  false,
				// This settings is used as a 'flag' and it does not matter what is set in the
				// state. If it is 'true' - then it means that 'update' procedure must set the
				// VM for customization at next boot and reboot it.
				DiffSuppressFunc: suppressFalse(),
				Description:      "'true' value will cause the VM to reboot on every 'apply' operation",
				ExactlyOneOf:     []string{"trigger_on_create", "trigger_on_every_apply"},
			},
			"ignore_existing_rule_error": {
				Type: schema.TypeBool,
				ValidateFunc: noopValueWarningValidator(true,
					"Using 'true' ignores error that is returned when autocreated Firewall and NAT rules already exist"),
				Optional:    true,
				Default:     false,
				Description: "Setting to 'true' will ignore error that is returned when rules are already created",
			},
		},
	}
}

func resourceVcdNsxtEdgeGatewayIpSpaceServicesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks are conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[ip space edge gateway services create] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[ip space edge gateway services create] error retrieving Edge Gateway: %s", err)
	}

	triggerOnCreate := d.Get("trigger_on_create").(bool)
	triggerOnEveryApply := d.Get("trigger_on_every_apply").(bool)
	ignoreExistingRuleError := d.Get("ignore_existing_rule_error").(bool)

	if triggerOnCreate || triggerOnEveryApply {
		err = nsxtEdge.ApplyIpSpaceDefaultServices()
		if err != nil {
			re := regexp.MustCompile(`Cannot apply default services on Gateway .* since it contains existing NAT rules`)
			if !ignoreExistingRuleError || (ignoreExistingRuleError && !re.MatchString(err.Error())) {
				return diag.Errorf("[ip space edge gateway services create] error applying IP Space default Gateway Services: %s", err)
			}
		}
	}

	d.SetId(nsxtEdge.EdgeGateway.ID)

	return resourceVcdNsxtEdgeGatewayIpSpaceServicesRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayIpSpaceServicesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	triggerOnEveryApply := d.Get("trigger_on_every_apply").(bool)
	ignoreExistingRuleError := d.Get("ignore_existing_rule_error").(bool)

	if !triggerOnEveryApply {
		return nil
	}

	// Handling locks are conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[ip space edge gateway services update] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[ip space edge gateway services update] error retrieving Edge Gateway: %s", err)
	}

	if triggerOnEveryApply {
		err = nsxtEdge.ApplyIpSpaceDefaultServices()
		if err != nil {
			re := regexp.MustCompile(`Cannot apply default services on Gateway .* since it contains existing NAT rules`)
			if !ignoreExistingRuleError || (ignoreExistingRuleError && !re.MatchString(err.Error())) {
				return diag.Errorf("[ip space edge gateway services create] error applying IP Space default Gateway Services: %s", err)
			}
		}
	}

	return resourceVcdNsxtEdgeGatewayIpSpaceServicesRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayIpSpaceServicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// message	"[ f93cc639-dd7f-42ac-8c39-1f4ecdd58160 ] Cannot apply default services on Gateway TestAccVcdIpSpacePublicEdgeGatewayDefaultServices since it contains existing NAT rules."

	// Explicitly setting 'trigger_on_every_apply' to false so that if
	dSet(d, "trigger_on_every_apply", false)
	return nil
}

func noOp(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
