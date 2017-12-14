package vcd

// This file is largly based on https://github.com/terraform-providers/terraform-provider-vsphere/blob/master/vsphere/internal/virtualdevice/virtual_machine_device_subresource.go
// by the vshpere provider team

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

// newSubresourceFunc is a method signature for the wrapper methods that create
// a new instance of a specific subresource  that is derived from the base
// subresoruce object. It's used in the general apply and read operation
// methods, which themselves are called usually from higher-level apply
// functions for virtual devices.
type newSubresourceFunc func(*VCDClient, int, int, *schema.ResourceData) SubresourceInstance

// SubresourceInstance is an interface for derivative objects of Subresource.
// It's used on the general apply and read operation methods, and contains both
// exported methods of the base Subresource type and the CRUD methods that
// should be supplied by derivative objects.
//
// Note that this interface should be used sparingly - as such, only the
// methods that are needed by inparticular functions external to most virtual
// device workflows are exported into this interface.
type SubresourceInstance interface {
	// Create(object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error)
	// Read(object.VirtualDeviceList) error
	// Update(object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error)
	// Delete(object.VirtualDeviceList) ([]types.BaseVirtualDeviceConfigSpec, error)

	Set(string, interface{}) error
	Schema() map[string]*schema.Schema
	State() map[string]interface{}
}

// resourceDataDiff is an interface comprised of functions common to both
// ResourceData and ResourceDiff.
//
// During any inparticular CRUD or diff alteration call, either one of
// ResourceData or ResourceDiff will be available. Both will never be available
// at the same time. Having these underlying values exposed directly presents a
// potentially unsafe API where one of them will be nil at any given time.
// Having this as an interface allows common behavior to be exposed directly,
// while still offering the ability to type assert in certain situations.
//
// This is not an exhaustive list of methods - any missing ones should be added
// as needed.
type resourceDataDiff interface {
	Id() string
	Get(string) interface{}
	HasChange(string) bool
}

// Subresource defines a common interface for device sub-resources in the
// vsphere_virtual_machine resource.
//
// This object is designed to be used by parts of the resource with workflows
// that are so complex in their own right that probably the only way to handle
// their management is to treat them like resources themselves.
//
// This structure of this resource loosely follows schema.Resource with having
// CRUD and maintaining a set of resource data to work off of. However, since
// we are using schema.Resource, we take some liberties that we normally would
// not be able to take, or need to take considering the context of the data we
// are working with.
//
// Inparticular functions implement this structure by creating an instance into
// it, much like how a resource creates itself by creating an instance of
// schema.Resource.
type Subresource struct {
	// The index of this subresource - should either be an index or hash. It's up
	// to the upstream object to set this to something useful.
	Index int

	// The resource schema. This is an internal field as we build on this field
	// later on with common keys for all subresources, namely the internal ID.
	schema map[string]*schema.Schema

	// The client connection.
	client *VCDClient

	// The resource data - this should be loaded when the resource is created.
	data map[string]interface{}

	// The old resource data, if it exists.
	olddata map[string]interface{}

	// Either a root-level ResourceData or ResourceDiff. The one that is
	// specifically present will depend on the context the Subresource is being
	// used in.
	rdd resourceDataDiff
}

// subresourceSchema is a map[string]*schema.Schema of common schema fields.
// This includes the internal_id field, which is used as a unique ID for the
// lifecycle of this resource.
func subresourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"key": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The unique device ID for this device within its virtual machine.",
		},
		"device_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The internally-computed address of this device, such as scsi:0:1, denoting scsi bus #0 and device unit 1.",
		},
	}
}

// Get hands off to r.data.Get, with an address relative to this subresource.
func (r *Subresource) Get(key string) interface{} {
	return r.data[key]
}

// Set sets the specified key/value pair in the subresource.
func (r *Subresource) Set(key string, value interface{}) {
	if v := NormalizeValue(value); v != nil {
		r.data[key] = v
	}
}

// HasChange checks to see if there has been a change in the resource data
// since the last update.
//
// Note that this operation may only be useful during update operations,
// depending on subresource-specific workflow.
func (r *Subresource) HasChange(key string) bool {
	o, n := r.GetChange(key)
	return !reflect.DeepEqual(o, n)
}

// GetChange gets the old and new values for the value specified by key.
func (r *Subresource) GetChange(key string) (interface{}, interface{}) {
	new := r.data[key]
	// No old data means no change,  so we use the new value as a placeholder.
	old := r.data[key]
	if r.olddata != nil {
		old = r.olddata[key]
	}
	return old, new
}

// GetWithVeto returns the value specified by key, but returns an error if it
// has changed. The intention here is to block changes to the resource in a
// fashion that would otherwise result in forcing a new resource.
func (r *Subresource) GetWithVeto(key string) (interface{}, error) {
	if r.HasChange(key) {
		old, new := r.GetChange(key)
		return r.Get(key), fmt.Errorf("cannot change the value of %q - (old: %v new: %v)", key, old, new)
	}
	return r.Get(key), nil
}

// Data returns the underlying data map.
func (r *Subresource) Data() map[string]interface{} {
	return r.data
}

// Hash calculates a set hash for the current data. If you want a hash for
// error reporting a device address, it's probably a good idea to run this at
// the beginning of a run as any set calls will change the value this
// ultimately calculates.
func (r *Subresource) Hash() int {
	hf := schema.HashResource(&schema.Resource{Schema: r.schema})
	return hf(r.data)
}

// subresourceListString takes a list of sub-resources and pretty-prints the
// key and device address.
func subresourceListString(data []interface{}) string {
	var strs []string
	for _, v := range data {
		if v == nil {
			strs = append(strs, "(<nil>)")
			continue
		}
		m := v.(map[string]interface{})
		devaddr := m["device_address"].(string)
		if devaddr == "" {
			devaddr = "<new device>"
		}
		strs = append(strs, fmt.Sprintf("(key %d at %s)", m["key"].(int), devaddr))
	}
	return strings.Join(strs, ",")
}
