package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdIpSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdIpSetCreate,
		Read:   resourceVcdIpSetRead,
		Update: resourceVcdIpSetUpdate,
		Delete: resourceVcdIpSetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdIpSetImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP set name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP set description",
			},
			"is_inheritance_allowed": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Allows visibility in underlying scopes (Default is true)",
			},
			"ip_addresses": {
				Required:    true,
				Type:        schema.TypeSet,
				Description: "A set of IP address, CIDR, IP range objects",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// resourceVcdIpSetCreate creates an IP set based on schema data
func resourceVcdIpSetCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating IP set with name %s", d.Get("name"))
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	ipSet, err := getIpSet(d, vdc)
	if err != nil {
		return fmt.Errorf("unable to make IP set query: %s", err)
	}

	createdIpSet, err := vdc.CreateNsxvIpSet(ipSet)
	if err != nil {
		return fmt.Errorf("error creating new IP set: %s", err)
	}

	log.Printf("[DEBUG] IP set with name %s created. Id: %s", createdIpSet.Name, createdIpSet.ID)
	d.SetId(createdIpSet.ID)
	return resourceVcdIpSetRead(d, meta)
}

// resourceVcdIpSetUpdate updates an IP set based on schema data
func resourceVcdIpSetUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating IP set with ID %s", d.Id())

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	ipSet, err := getIpSet(d, vdc)
	if err != nil {
		return fmt.Errorf("unable to make IP set query: %s", err)
	}
	ipSet.ID = d.Id() // ID is needed to update IP set

	_, err = vdc.UpdateNsxvIpSet(ipSet)
	if err != nil {
		return fmt.Errorf("error updating IP set with ID %s: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Updated IP set with ID %s", d.Id())
	return resourceVcdIpSetRead(d, meta)
}

func datasourceVcdIpSetRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdIpSetRead(d, meta, "datasource")
}

func resourceVcdIpSetRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdIpSetRead(d, meta, "resource")
}

// genericVcdIpSetRead reads all data and persists it on statefile.
// When "origin" == "datasource" it will search for IP set by name and use d.SetId
// When "origin" != "datasource" it will search for IP set by ID and do not perform d.SetId
func genericVcdIpSetRead(d *schema.ResourceData, meta interface{}, origin string) error {
	log.Printf("[DEBUG] Reading IP set with ID %s", d.Id())
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	var ipSet *types.EdgeIpSet

	// Only if it is a data source we will find IP set by name, otherwise we always have an ID
	if origin == "datasource" {
		ipSet, err = vdc.GetNsxvIpSetByName(d.Get("name").(string))
	} else {
		ipSet, err = vdc.GetNsxvIpSetById(d.Id())
	}

	if govcd.IsNotFound(err) && origin == "resource" {
		log.Printf("[INFO] unable to find IP set with ID %s: %s. Removing from state", d.Id(), err)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("unable to find IP set with ID %s: %s", d.Id(), err)
	}

	err = setIpSetData(d, ipSet, vdc, origin)
	if err != nil {
		return fmt.Errorf("unable to store data in statefile: %s", err)
	}

	if origin == "datasource" {
		d.SetId(ipSet.ID)
	}

	log.Printf("[DEBUG] Read IP set with ID %s", d.Id())
	return nil
}

// resourceVcdIpSetDelete delete IP set based on its ID
func resourceVcdIpSetDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting IP set with ID %s", d.Id())
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	err = vdc.DeleteNsxvIpSetById(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting IP set with id %s: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleted IP set with ID %s", d.Id())
	d.SetId("")
	return nil
}

// resourceVcdIpSetImport
func resourceVcdIpSetImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified in such way org-name.vdc-name.ipset-name")
	}
	orgName, vdcName, ipSetName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	ipSet, err := vdc.GetNsxvIpSetByName(ipSetName)

	if err != nil {
		return nil, fmt.Errorf("unable to find IP set with name %s", ipSetName)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.SetId(ipSet.ID)

	return []*schema.ResourceData{d}, nil
}

// getIpSet convert terraform schema definition and creates a *types.EdgeIpSet
func getIpSet(d *schema.ResourceData, vdc *govcd.Vdc) (*types.EdgeIpSet, error) {

	ipSet := &types.EdgeIpSet{
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		InheritanceAllowed: takeBoolPointer(d.Get("is_inheritance_allowed").(bool)),
	}

	ipAddressesSlice := convertSchemaSetToSliceOfStrings(d.Get("ip_addresses").(*schema.Set))
	ipSet.IPAddresses = strings.Join(ipAddressesSlice, ",")

	return ipSet, nil
}

// setIpSetData sets data into statefile using a provided IP set type
func setIpSetData(d *schema.ResourceData, ipSet *types.EdgeIpSet, vdc *govcd.Vdc, origin string) error {

	if origin == "resource" {
		d.Set("name", ipSet.Name)
	}

	d.Set("description", ipSet.Description)
	d.Set("is_inheritance_allowed", ipSet.InheritanceAllowed)

	// convert comma separated list of ip addresses to TypeSet and set it
	var ipAddressesSlice []interface{}
	ipSlice := strings.Split(ipSet.IPAddresses, ",")
	if len(ipSlice) > 0 {
		ipAddressesSlice = make([]interface{}, len(ipSlice))
		for ipIndex, ipAddress := range ipSlice {
			ipAddressesSlice[ipIndex] = ipAddress
		}
	}
	ipAddressSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), ipAddressesSlice)

	err := d.Set("ip_addresses", ipAddressSet)
	if err != nil {
		return fmt.Errorf("could not convert ip_addresses to set: %s", err)
	}

	return nil
}

// ipSetIdsToNames looks up IP sets by IDs and returns list of their names
func ipSetIdsToNames(ipSetIds []string, vdc *govcd.Vdc) ([]string, error) {
	ipSetNames := make([]string, len(ipSetIds))

	allIpSets, err := vdc.GetAllNsxvIpSets()
	// If no IP sets are found - return empty list of names
	if govcd.IsNotFound(err) {
		return ipSetIds, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to fetch all IP sets in vDC %s: %s", vdc.Vdc.Name, err)
	}

	for index, ipSetId := range ipSetIds {
		var ipSetFound bool
		for _, ipSet := range allIpSets {
			if ipSet.ID == ipSetId {
				ipSetNames[index] = ipSet.Name
				ipSetFound = true
			}
		}
		// If ID was not found - fail early
		if !ipSetFound {
			return nil, fmt.Errorf("could not find IP set with ID %s", ipSetId)
		}
	}

	return ipSetNames, nil
}

// ipSetNamesToIds looks up IP set names by their IDs
func ipSetNamesToIds(ipSetNames []string, vdc *govcd.Vdc, isShortFormat bool) ([]string, error) {
	ipSetIds := make([]string, len(ipSetNames))

	// When no names are passed - there is no need to lookup IP sets
	if len(ipSetNames) == 0 {
		return ipSetIds, nil
	}

	allIpSets, err := vdc.GetAllNsxvIpSets()
	// If no IP sets are found in vCD - return empty list of IDs
	if govcd.IsNotFound(err) {
		return ipSetIds, nil
	}

	if err != nil {
		return nil, fmt.Errorf("unable to fetch all IP sets in vDC %s: %s", vdc.Vdc.Name, err)
	}

	for index, ipSetName := range ipSetNames {
		var ipSetFound bool
		for _, ipSet := range allIpSets {
			if ipSet.Name == ipSetName {
				if isShortFormat {
					ipSetIds[index] = strings.Split(ipSet.ID, ":")[1]
				} else {
					ipSetIds[index] = ipSet.ID
				}
				ipSetFound = true
			}
		}
		// If ID was not found - fail early
		if !ipSetFound {
			return nil, fmt.Errorf("could not find IP set with Name %s", ipSetName)
		}
	}

	return ipSetIds, nil
}
