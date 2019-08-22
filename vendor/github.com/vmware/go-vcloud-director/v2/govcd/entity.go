/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

type genericGetter func(string, bool) (interface{}, error)

// getEntityByNameOrId finds a generic entity by Name Or ID
// On success, returns an empty interface representing a pointer to the structure and a nil error
// On failure, returns a nil pointer and an error
// Example usage:
//
// func (org *Org) GetCatalogByNameOrId(identifier string, refresh bool) (*Catalog, error) {
// 	getByName := func(name string, refresh bool) (interface{}, error) {
// 		return org.GetCatalogByName(name, refresh)
// 	}
// 	getById := func(id string, refresh bool) (interface{}, error) {
// 	  return org.GetCatalogById(id, refresh)
// 	}
// 	entity, err := getEntityByNameOrId(getByName, getById, identifier, refresh)
//  if entity != nil {
//    return nil, err
//  }
// 	return entity.(*Catalog), err
// }
func getEntityByNameOrId(getByName, getById genericGetter, identifier string, refresh bool) (interface{}, error) {

	var byNameErr, byIdErr error
	var entity interface{}

	entity, byIdErr = getById(identifier, refresh)
	if byIdErr == nil {
		// Found by ID
		return entity, nil
	}
	if IsNotFound(byIdErr) {
		// Not found by ID, try by name
		entity, byNameErr = getByName(identifier, false)
		return entity, byNameErr
	} else {
		// On any other error, we return it
		return nil, byIdErr
	}
}
