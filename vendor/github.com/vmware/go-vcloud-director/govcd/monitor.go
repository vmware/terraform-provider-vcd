/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// Contains auxiliary functions to show library entities structure.
// Used for debugging and testing.
package govcd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
)

// For each library {entity}, we have two functions: Show{Entity} and Log{Entity}
// The first one shows the contents of the entity on screen
// The second one logs into the default util.Logger
// Both functions use JSON as the entity format

// Available entities:
// org
// adminOrg
// vdc
// catalog
// catalogItem
// adminCatalog
// network
// externalNetwork
// vapp
// task
// Edge Gateway service configuration

func out(destination, format string, args ...interface{}) {
	switch destination {
	case "screen":
		fmt.Printf(format, args...)
	case "log":
		util.Logger.Printf(format, args...)
	default:
		fmt.Printf("Unhandled destination: %s\n", destination)
	}
}

// Returns a vApp structure as JSON
func prettyVapp(vapp types.VApp) string {
	byteBuf, err := json.MarshalIndent(vapp, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns a VDC structure as JSON
func prettyVdc(vdc types.Vdc) string {
	byteBuf, err := json.MarshalIndent(vdc, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns a Catalog Item structure as JSON
func prettyCatalogItem(catalogItem types.CatalogItem) string {
	byteBuf, err := json.MarshalIndent(catalogItem, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns a Catalog structure as JSON
func prettyCatalog(catalog types.Catalog) string {
	byteBuf, err := json.MarshalIndent(catalog, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns an Admin Catalog structure as JSON
func prettyAdminCatalog(catalog types.AdminCatalog) string {
	byteBuf, err := json.MarshalIndent(catalog, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns an Org structure as JSON
func prettyOrg(org types.Org) string {
	byteBuf, err := json.MarshalIndent(org, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns an Admin Org structure as JSON
func prettyAdminOrg(org types.AdminOrg) string {
	byteBuf, err := json.MarshalIndent(org, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns a Disk structure as JSON
func prettyDisk(disk types.Disk) string {
	byteBuf, err := json.MarshalIndent(disk, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns an External Network structure as JSON
func prettyExternalNetwork(network types.ExternalNetworkReference) string {
	byteBuf, err := json.MarshalIndent(network, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns a Network structure as JSON
func prettyNetworkConf(conf types.OrgVDCNetwork) string {
	byteBuf, err := json.MarshalIndent(conf, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns a Task structure as JSON
func prettyTask(task *types.Task) string {
	byteBuf, err := json.MarshalIndent(task, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

// Returns an Edge Gateway service configuration structure as JSON
func prettyEdgeGatewayServiceConfiguration(conf *types.EdgeGatewayServiceConfiguration) string {
	byteBuf, err := json.MarshalIndent(conf, " ", " ")
	if err == nil {
		return fmt.Sprintf("%s\n", string(byteBuf))
	}
	return ""
}

func LogNetwork(conf types.OrgVDCNetwork) {
	out("log", prettyNetworkConf(conf))
}

func ShowNetwork(conf types.OrgVDCNetwork) {
	out("screen", prettyNetworkConf(conf))
}

func LogExternalNetwork(network types.ExternalNetworkReference) {
	out("log", prettyExternalNetwork(network))
}

func ShowExternalNetwork(network types.ExternalNetworkReference) {
	out("screen", prettyExternalNetwork(network))
}

func LogVapp(vapp types.VApp) {
	out("log", prettyVapp(vapp))
}

func ShowVapp(vapp types.VApp) {
	out("screen", prettyVapp(vapp))
}

func ShowOrg(org types.Org) {
	out("screen", prettyOrg(org))
}

func LogOrg(org types.Org) {
	out("log", prettyOrg(org))
}

func ShowAdminOrg(org types.AdminOrg) {
	out("screen", prettyAdminOrg(org))
}

func LogAdminOrg(org types.AdminOrg) {
	out("log", prettyAdminOrg(org))
}

func ShowVdc(vdc types.Vdc) {
	out("screen", prettyVdc(vdc))
}

func LogVdc(vdc types.Vdc) {
	out("log", prettyVdc(vdc))
}

func ShowDisk(disk types.Disk) {
	out("screen", prettyDisk(disk))
}

func LogDisk(disk types.Disk) {
	out("log", prettyDisk(disk))
}
func ShowCatalog(catalog types.Catalog) {
	out("screen", prettyCatalog(catalog))
}

func LogCatalog(catalog types.Catalog) {
	out("log", prettyCatalog(catalog))
}

func ShowCatalogItem(catalogItem types.CatalogItem) {
	out("screen", prettyCatalogItem(catalogItem))
}

func LogCatalogItem(catalogItem types.CatalogItem) {
	out("log", prettyCatalogItem(catalogItem))
}

func ShowAdminCatalog(catalog types.AdminCatalog) {
	out("screen", prettyAdminCatalog(catalog))
}

func LogAdminCatalog(catalog types.AdminCatalog) {
	out("log", prettyAdminCatalog(catalog))
}

// Auxiliary function to monitor a task
// It can be used in association with WaitInspectTaskCompletion
func outTask(destination string, task *types.Task, howManyTimes int, elapsed time.Duration, first, last bool) {
	if task == nil {
		out(destination, "Task is null\n")
		return
	}
	out(destination, prettyTask(task))

	out(destination, "progress: [%s:%d] %d%%\n", elapsed.Round(1*time.Second), howManyTimes, task.Progress)
	out(destination, "-------------------------------\n")
}

func LogTask(task *types.Task, howManyTimes int, elapsed time.Duration, first, last bool) {
	outTask("log", task, howManyTimes, elapsed, first, last)
}

func ShowTask(task *types.Task, howManyTimes int, elapsed time.Duration, first, last bool) {
	outTask("screen", task, howManyTimes, elapsed, first, last)
}
