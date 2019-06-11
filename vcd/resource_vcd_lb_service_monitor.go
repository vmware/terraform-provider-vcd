package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdLbServiceMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdLbServiceMonitorCreate,
		Read:   resourceVcdLbServiceMonitorRead,
		Update: resourceVcdLbServiceMonitorUpdate,
		Delete: resourceVcdLbServiceMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdLbServiceMonitorImport,
		},

		Schema: map[string]*schema.Schema{
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique Service Monitor name",
			},
			"interval": &schema.Schema{
				Type:        schema.TypeInt,
				Default:     10,
				Optional:    true,
				Description: "Interval at which a server is to be monitored",
			},
			"timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Default:     15,
				Optional:    true,
				Description: "Maximum time in seconds within which a response from the server must be received",
			},
			"max_retries": &schema.Schema{
				Type:        schema.TypeInt,
				Default:     3,
				Optional:    true,
				Description: "Number of times the specified monitoring Method must fail sequentially before the server is declared down",
			},
			"type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Way in which you want to send the health check request to the server. One of HTTP, HTTPS, TCP, ICMP, or UDP",
				ValidateFunc: validateCase("lower"),
			},
			"expected": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "String that the monitor expects to match in the status line of the HTTP or HTTPS response (for example, HTTP/1.1)",
			},
			"method": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Method to be used to detect server status. One of OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, or CONNECT",
				ValidateFunc: validateCase("upper"),
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL to be used in the server status request",
			},
			"send": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Data to be sent",
			},
			"receive": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "String to be matched in the response content",
			},
			"extension": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Advanced monitor parameters as key=value pairs",
			},
		},
	}
}

// validateCase checks if a string is of caseType "upper" or "lower"
func validateCase(caseType string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		switch caseType {
		case "upper":
			if strings.ToUpper(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be upper cased, got: %s", v))
			}
		case "lower":
			if strings.ToLower(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be lower cased, got: %s", v))
			}
		default:
			panic("unsupported validation type for validateCase() function")
		}
		return
	}
}

func resourceVcdLbServiceMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	lbMonitor, err := expandLBMonitor(d)
	if err != nil {
		return fmt.Errorf("unable to expand load balancer service monitor: %s", err)
	}

	createdMonitor, err := edgeGateway.CreateLBServiceMonitor(lbMonitor)
	if err != nil {
		return fmt.Errorf("error creating new load balancer service monitor: %s", err)
	}

	d.SetId(createdMonitor.ID)
	return nil
}

func resourceVcdLbServiceMonitorRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBMonitor, err := edgeGateway.ReadLBServiceMonitor(&types.LBMonitor{ID: d.Id()})
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer service monitor with ID %s: %s", d.Id(), err)
	}

	return flattenLBMonitor(d, readLBMonitor)
}

func resourceVcdLbServiceMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateLBMonitorConfig, err := expandLBMonitor(d)
	if err != nil {
		return fmt.Errorf("could not expand monitor for update: %s", err)
	}

	updatedLBMonitor, err := edgeGateway.UpdateLBServiceMonitor(updateLBMonitorConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer service monitor with ID %s: %s", d.Id(), err)
	}

	return flattenLBMonitor(d, updatedLBMonitor)
}

func resourceVcdLbServiceMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.DeleteLBServiceMonitor(&types.LBMonitor{ID: d.Id()})
	if err != nil {
		return fmt.Errorf("error deleting load balancer service monitor: %s", err)
	}

	d.SetId("")
	return nil
}

// resourceVcdLbServiceMonitorImport expects dot formatted path to Load Balancer Service Monitor
// i.e. my-org.my-org-vdc.my-edge-gw.existing-sm
func resourceVcdLbServiceMonitorImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified in such way my-org.my-org-vdc.my-edge-gw.existing-lb-service-monitor-name")
	}
	orgName, vdcName, edgeName, monitorName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBMonitor, err := edgeGateway.ReadLBServiceMonitor(&types.LBMonitor{Name: monitorName})
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find load balancer service monitor with ID %s: %s", d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)
	d.Set("name", monitorName)

	d.SetId(readLBMonitor.ID)
	return []*schema.ResourceData{d}, nil
}

func expandLBMonitor(d *schema.ResourceData) (*types.LBMonitor, error) {
	lbMonitor := &types.LBMonitor{
		Name:       d.Get("name").(string),
		Interval:   d.Get("interval").(int),
		Timeout:    d.Get("timeout").(int),
		Type:       d.Get("type").(string),
		MaxRetries: d.Get("max_retries").(int),
		Expected:   d.Get("expected").(string),
		Method:     d.Get("method").(string),
		URL:        d.Get("url").(string),
		Send:       d.Get("send").(string),
		Receive:    d.Get("receive").(string),
		Extension:  expandLBMonitorExtension(d),
	}

	return lbMonitor, nil
}

// expandLBMonitorExtension expands the specified map for sending via API. It appends newline to every extension as
// per API requirement. Based on the research the underlying structure should not cause problems because duplicate keys
// are not needed and order of the keys does not matter for API.
// Example API call string for Extension field:
// <extension>delay=2
// critical=3
// escape</extension>
func expandLBMonitorExtension(d *schema.ResourceData) string {
	var extensionString string
	extension := d.Get("extension").(map[string]interface{})
	for k, v := range extension {
		if k != "" && v != "" { // When key and value are given it must look like "content-type=STRING"
			extensionString += k + "=" + v.(string) + "\n"
		} else { // If only key is specified it does not need equals sign. Like "no-body" extension
			extensionString += k + "\n"
		}
	}
	return extensionString
}

func flattenLBMonitor(d *schema.ResourceData, lBmonitor *types.LBMonitor) error {
	d.Set("interval", lBmonitor.Interval)
	d.Set("timeout", lBmonitor.Timeout)
	d.Set("max_retries", lBmonitor.MaxRetries)
	d.Set("type", lBmonitor.Type)
	// Optional attributes may not necessarilly
	d.Set("method", lBmonitor.Method)
	d.Set("url", lBmonitor.URL)
	d.Set("send", lBmonitor.Send)
	d.Set("receive", lBmonitor.Receive)
	d.Set("expected", lBmonitor.Expected)

	if err := flattenLBMonitorExtension(d, lBmonitor); err != nil {
		return err
	}

	return nil
}

// flattenLBMonitorExtension is responsive to parse response extension field from API and
// store it in the map. It supports flattening `key=value` or `key` notations. Each of them must be
// separated by newline.
func flattenLBMonitorExtension(d *schema.ResourceData, lBmonitor *types.LBMonitor) error {
	extensionStorage := make(map[string]string)

	if lBmonitor.Extension != "" {
		kvList := strings.Split(lBmonitor.Extension, "\n")
		for _, extensionLine := range kvList {
			// Skip empty lines
			if extensionLine == "" {
				continue
			}

			// When key=extensionLine format is present
			if strings.Contains(extensionLine, "=") {
				keyValue := strings.Split(extensionLine, "=")
				if len(keyValue) != 2 {
					return fmt.Errorf("unable to flatten extension field %s", extensionLine)
				}
				// Populate extension data with key value
				extensionStorage[keyValue[0]] = keyValue[1]
				// If there was no "=" sign then it means whole line is just key. Like `no-body`, `linespan`
			} else {
				extensionStorage[extensionLine] = ""
			}
		}

	}

	d.Set("extension", extensionStorage)
	return nil
}
