/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// CreateNsxvIpSet creates an IP set from *types.EdgeIpSet. IP set defines a group of IP addresses
// that you can add as the source or destination in a firewall rule or in DHCP relay configuration.
func (vdc *Vdc) CreateNsxvIpSet(ipSetConfig *types.EdgeIpSet) (*types.EdgeIpSet, error) {
	if err := validateCreateNsxvIpSet(ipSetConfig); err != nil {
		return nil, err
	}

	vdcId, err := getUuidFromHref(vdc.Vdc.HREF)
	if err != nil {
		return nil, fmt.Errorf("unable to get vdc ID from HREF: %s", err)
	}

	// build a path for IP set creation. The endpoint should look like:
	// https://_hostname_/network/services/ipset/f9daf2da-b4f9-4921-a2f4-d77a943a381c where the
	// trailing UUID is vDC ID
	httpPath, err := vdc.buildNsxvNetworkServiceEndpointURL(types.NsxvIpSetServicePath + "/" + vdcId)
	if err != nil {
		return nil, fmt.Errorf("could not get network services API endpoint for IP set: %s", err)
	}

	// Success or an error of type types.NSXError is expected
	_, err = vdc.client.ExecuteParamRequestWithCustomError(httpPath, nil, http.MethodPost, types.AnyXMLMime,
		"error creating IP set: %s", ipSetConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	createdIpSet, err := vdc.GetNsxvIpSetByName(ipSetConfig.Name)
	if err != nil {
		return nil, fmt.Errorf("could not lookup newly created IP set with name %s: %s", ipSetConfig.Name, err)
	}

	return createdIpSet, nil
}

// UpdateNsxvIpSet sends all fields of ipSetConfig. Omiting a value may reset it. ID is mandatory to
// perform update.
// Because the API always requires a Revision to be sent - the update fetches latest revision number
// automatically and embeds into the update structure.
func (vdc *Vdc) UpdateNsxvIpSet(ipSetConfig *types.EdgeIpSet) (*types.EdgeIpSet, error) {
	err := validateUpdateNsxvIpSet(ipSetConfig)
	if err != nil {
		return nil, err
	}

	// Inject latest Revision for this IP set so that API accepts change
	currentIpSet, err := vdc.GetNsxvIpSetById(ipSetConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch current IP set: %s", err)
	}
	ipSetConfig.Revision = currentIpSet.Revision

	httpPath, err := vdc.buildNsxvNetworkServiceEndpointURL(types.NsxvIpSetServicePath + "/" + ipSetConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get network services API endpoint for IP set: %s", err)
	}

	// Result is either 204 for success, or an error of type types.NSXError
	errString := fmt.Sprintf("error while updating IP set with ID %s :%%s", ipSetConfig.ID)
	_, err = vdc.client.ExecuteRequestWithCustomError(httpPath, http.MethodPut, types.AnyXMLMime,
		errString, ipSetConfig, &types.NSXError{})
	if err != nil {
		return nil, err
	}

	updatedIpSet, err := vdc.GetNsxvIpSetById(ipSetConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("could not lookup updated IP set with ID %s: %s", ipSetConfig.ID, err)
	}

	return updatedIpSet, nil
}

// GetNsxvIpSetByName searches for IP set by name. Names are unique therefore it can find only one.
// Returns ErrorEntityNotFound if an IP set is not found
func (vdc *Vdc) GetNsxvIpSetByName(name string) (*types.EdgeIpSet, error) {
	if err := validateGetNsxvIpSet("", name); err != nil {
		return nil, err
	}

	allIpSets, err := vdc.GetAllNsxvIpSets()
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("[DEBUG] Searching for IP set with name: %s", name)
	for _, ipSet := range allIpSets {
		util.Logger.Printf("[DEBUG] Checking IP set: %#+v", ipSet)
		if ipSet.Name != "" && ipSet.Name == name {
			return ipSet, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetNsxvIpSetById searches for IP set by ID. Returns ErrorEntityNotFound if an IP set is not found
func (vdc *Vdc) GetNsxvIpSetById(id string) (*types.EdgeIpSet, error) {
	if err := validateGetNsxvIpSet(id, ""); err != nil {
		return nil, err
	}

	allIpSets, err := vdc.GetAllNsxvIpSets()
	if err != nil {
		return nil, err
	}

	util.Logger.Printf("[DEBUG] Searching for IP set with id: %s", id)
	for _, ipSet := range allIpSets {
		util.Logger.Printf("[DEBUG] Checking IP set: %#+v", ipSet)
		if ipSet.ID != "" && ipSet.ID == id {
			return ipSet, nil
		}
	}

	return nil, ErrorEntityNotFound
}

// GetNsxvIpSetByNameOrId uses the same identifier to search by name and by ID. Priority is to try
// and find the IP set by ID. If it is not found - then a search by name is performed.
func (vdc *Vdc) GetNsxvIpSetByNameOrId(identifier string) (*types.EdgeIpSet, error) {
	getByName := func(name string, refresh bool) (interface{}, error) { return vdc.GetNsxvIpSetByName(name) }
	getById := func(id string, refresh bool) (interface{}, error) { return vdc.GetNsxvIpSetById(id) }
	entity, err := getEntityByNameOrId(getByName, getById, identifier, true)
	if entity == nil {
		return nil, err
	}
	return entity.(*types.EdgeIpSet), err
}

// GetAllNsxvIpSets retrieves all IP sets and returns []*types.EdgeIpSet or an
// error of type ErrorEntityNotFound if there are no IP sets
func (vdc *Vdc) GetAllNsxvIpSets() ([]*types.EdgeIpSet, error) {
	vdcId, err := getUuidFromHref(vdc.Vdc.HREF)
	if err != nil {
		return nil, fmt.Errorf("unable to get vdc ID from HREF: %s", err)
	}

	// build a path for to read all IP sets in a scope. A scope is defined by vDC ID. The endpoint
	// should look like:
	// https://192.168.1.109/network/services/ipset/scope/f9daf2da-b4f9-4921-a2f4-d77a943a381c where
	// the trailing UUID is vDC ID
	httpPath, err := vdc.buildNsxvNetworkServiceEndpointURL(types.NsxvIpSetServicePath + "/scope/" + vdcId)
	if err != nil {
		return nil, fmt.Errorf("could not get network services API endpoint for IP set: %s", err)
	}

	// Anonymous struct to unwrap list of IP sets <list><ipset></ipset><ipset></ipset></list>
	ipSetsResponse := &struct {
		XMLName          xml.Name `xml:"list"`
		types.EdgeIpSets `xml:"ipset"`
	}{}

	// This query returns all IP sets on the scope (scoped by vDC ID)
	errString := fmt.Sprintf("unable to read IP sets for scope %s: %%s", vdcId)
	_, err = vdc.client.ExecuteRequest(httpPath, http.MethodGet, types.AnyXMLMime, errString, nil, ipSetsResponse)
	if err != nil {
		return nil, err
	}

	if len(ipSetsResponse.EdgeIpSets) == 0 {
		return nil, ErrorEntityNotFound
	}

	return ipSetsResponse.EdgeIpSets, nil
}

// DeleteNsxvIpSetById deletes IP set by its ID which is formatted as
// f9daf2da-b4f9-4921-a2f4-d77a943a381c:ipset-9
func (vdc *Vdc) DeleteNsxvIpSetById(id string) error {
	err := validateDeleteNsxvIpSet(id, "")
	if err != nil {
		return err
	}

	// build a path for to delete exact IP set sample path is: DELETE API-URL/services/ipset/id:ipset-#
	// https://192.168.1.109/network/services/ipset/f9daf2da-b4f9-4921-a2f4-d77a943a381c:ipset-9
	httpPath, err := vdc.buildNsxvNetworkServiceEndpointURL(types.NsxvIpSetServicePath + "/" + id)
	if err != nil {
		return fmt.Errorf("could not get network services API endpoint for IP set: %s", err)
	}

	errString := fmt.Sprintf("unable to delete IP set with ID %s: %%s", id)
	_, err = vdc.client.ExecuteRequestWithCustomError(httpPath, http.MethodDelete, types.AnyXMLMime,
		errString, nil, &types.NSXError{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteNsxvIpSetById deletes IP set by its name
func (vdc *Vdc) DeleteNsxvIpSetByName(name string) error {
	err := validateDeleteNsxvIpSet("", name)
	if err != nil {
		return err
	}

	// Get IP set by name
	ipSet, err := vdc.GetNsxvIpSetByName(name)
	if err != nil {
		return err
	}

	return vdc.DeleteNsxvIpSetById(ipSet.ID)
}

func validateCreateNsxvIpSet(ipSetConfig *types.EdgeIpSet) error {

	if ipSetConfig.Name == "" {
		return fmt.Errorf("IP set must have name defined")
	}

	if ipSetConfig.IPAddresses == "" {
		return fmt.Errorf("IP set must IP addresses defined")
	}

	return nil
}

func validateUpdateNsxvIpSet(ipSetConfig *types.EdgeIpSet) error {

	if ipSetConfig.ID == "" {
		return fmt.Errorf("IP set ID must be set for update")
	}

	return validateCreateNsxvIpSet(ipSetConfig)
}

func validateGetNsxvIpSet(id, name string) error {
	if id == "" && name == "" {
		return fmt.Errorf("at least name or ID must be provided")
	}

	return nil
}

func validateDeleteNsxvIpSet(id, name string) error {
	return validateGetNsxvIpSet(id, name)
}
