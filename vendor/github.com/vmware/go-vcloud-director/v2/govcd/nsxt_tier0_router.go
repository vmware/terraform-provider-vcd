/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// NsxtTier0Router
type NsxtTier0Router struct {
	NsxtTier0Router *types.NsxtTier0Router
	client          *Client
}

// GetImportableNsxtTier0RouterByName retrieves NSX-T tier 0 router by given parent NSX-T manager ID and Tier 0 router
// name
//
// Warning. The API returns only unused Tier-0 routers (the ones that are not used in external networks yet)
//
// Note. NSX-T manager ID is mandatory and must be in URN format (e.g.
// urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc)

func (vcdCli *VCDClient) GetImportableNsxtTier0RouterByName(name, nsxtManagerId string) (*NsxtTier0Router, error) {
	if nsxtManagerId == "" {
		return nil, fmt.Errorf("no NSX-T manager ID specified")
	}

	if !isUrn(nsxtManagerId) {
		return nil, fmt.Errorf("NSX-T manager ID is not URN (e.g. 'urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc)', got: %s", nsxtManagerId)
	}

	if name == "" {
		return nil, fmt.Errorf("empty Tier 0 router name specified")
	}

	// Ideally FIQL filter could be used to filter on server side and get only desired result, but filtering on
	// 'displayName' is not yet supported. The only supported field for filtering is
	// _context==urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc to specify parent NSX-T manager (This
	// automatically happens in GetAllImportableNsxtTier0Routers()). The below filter injection is left as documentation.
	/*
		queryParameters := copyOrNewUrlValues(nil)
		queryParameters.Add("filter", "displayName=="+name)
	*/

	nsxtTier0Routers, err := vcdCli.GetAllImportableNsxtTier0Routers(nsxtManagerId, nil)
	if err != nil {
		return nil, fmt.Errorf("could not find NSX-T Tier-0 router with name '%s' for NSX-T manager with id '%s': %s",
			name, nsxtManagerId, err)
	}

	// TODO remove this when FIQL supports filtering on 'displayName'
	nsxtTier0Routers = filterNsxtTier0RoutersInExternalNetworks(name, nsxtTier0Routers)
	// EOF TODO remove this when FIQL supports filtering on 'displayName'

	if len(nsxtTier0Routers) == 0 {
		// ErrorEntityNotFound is injected here for the ability to validate problem using ContainsNotFound()
		return nil, fmt.Errorf("%s: no NSX-T Tier-0 router with name '%s' for NSX-T manager with id '%s' found",
			ErrorEntityNotFound, name, nsxtManagerId)
	}

	if len(nsxtTier0Routers) > 1 {
		return nil, fmt.Errorf("more than one (%d) NSX-T Tier-0 router with name '%s' for NSX-T manager with id '%s' found",
			len(nsxtTier0Routers), name, nsxtManagerId)
	}

	return nsxtTier0Routers[0], nil
}

// filterNsxtTier0RoutersInExternalNetworks is created as a fix for local filtering instead of using
// FIQL filter (because it does not support it).
func filterNsxtTier0RoutersInExternalNetworks(name string, allNnsxtTier0Routers []*NsxtTier0Router) []*NsxtTier0Router {
	filteredNsxtTier0Routers := make([]*NsxtTier0Router, 0)
	for index, nsxtTier0Router := range allNnsxtTier0Routers {
		if allNnsxtTier0Routers[index].NsxtTier0Router.DisplayName == name {
			filteredNsxtTier0Routers = append(filteredNsxtTier0Routers, nsxtTier0Router)
		}
	}

	return filteredNsxtTier0Routers

}

// GetAllImportableNsxtTier0Routers retrieves all NSX-T Tier-0 routers using OpenAPI endpoint. Query parameters can be
// supplied to perform additional filtering. By default it injects FIQL filter _context==nsxtManagerId (e.g.
// _context==urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc) because it is mandatory to list child Tier-0
// routers.
//
// Warning. The API returns only unused Tier-0 routers (the ones that are not used in external networks yet)
//
// Note. IDs of Tier-0 routers do not have a standard and may look as strings when they are created using UI or as UUIDs
// when they are created using API
func (vcdCli *VCDClient) GetAllImportableNsxtTier0Routers(nsxtManagerId string, queryParameters url.Values) ([]*NsxtTier0Router, error) {
	if !isUrn(nsxtManagerId) {
		return nil, fmt.Errorf("NSX-T manager ID is not URN (e.g. 'urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc)', got: %s", nsxtManagerId)
	}

	endpoint := types.OpenApiPathVersion1_0_0 + types.OpenApiEndpointImportableTier0Routers
	minimumApiVersion, err := vcdCli.Client.checkOpenApiEndpointCompatibility(endpoint)
	if err != nil {
		return nil, err
	}

	urlRef, err := vcdCli.Client.OpenApiBuildEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	// Get all Tier-0 routers that are accessible to an organization VDC. Routers that are already associated with an
	// External Network are filtered out. The “_context” filter key must be set with the id of the NSX-T manager for which
	// we want to get the Tier-0 routers for.
	//
	// _context==urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc

	// Create a copy of queryParameters so that original queryParameters are not mutated (because a map is always a
	// reference)
	queryParams := queryParameterFilterAnd("_context=="+nsxtManagerId, queryParameters)

	typeResponses := []*types.NsxtTier0Router{{}}
	err = vcdCli.Client.OpenApiGetAllItems(minimumApiVersion, urlRef, queryParams, &typeResponses)
	if err != nil {
		return nil, err
	}

	returnObjects := make([]*NsxtTier0Router, len(typeResponses))
	for sliceIndex := range typeResponses {
		returnObjects[sliceIndex] = &NsxtTier0Router{
			NsxtTier0Router: typeResponses[sliceIndex],
			client:          &vcdCli.Client,
		}
	}

	return returnObjects, nil
}
