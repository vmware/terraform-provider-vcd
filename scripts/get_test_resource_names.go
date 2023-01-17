// get_test_resource_names collects the resource names from the test files in test-artifacts
// By default, it skips the names starting with 'Test' or 'test', as they are otherwise
// recognised and handled during the leftover removal operation.
// This file gets copied to ./vcd/test-artifacts when we run `make test-binary` or `make test-binary-prepare`
//
// Run the code in ./vcd/test-artifacts, possibly after running `make test-binary-prepare` in the top directory
// go run get_test_resource_names.go | vim -
//
// At the end, this program prints two things:
// 1: the JSON text of our entities (to be used stand-alone)
// 2: the Go representation of the entities (to be pasted in remove_leftovers_test.go)

// Version 0.1 - 2022-12-14

package main

import (
	"encoding/json"
	"fmt"
	"github.com/kr/pretty"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type entityDef struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Comment string `json:"comment,omitempty"`
}

// entityList is a collection of entityDef
type entityList []entityDef

// typeReplacements replaces the found type with a more generic one that
// is supported in remove-leftover tool.
// When a type is not supported at all, the replacement is an empty string
var typeReplacements = map[string]string{
	"vcd_network_routed_v2":                   "vcd_network",
	"vcd_network_direct_v2":                   "vcd_network",
	"vcd_network_isolated_v2":                 "vcd_network",
	"vcd_network_routed":                      "vcd_network",
	"vcd_network_direct":                      "vcd_network",
	"vcd_network_isolated":                    "vcd_network",
	"vcd_subscribed_catalog":                  "vcd_catalog",
	"vcd_catalog_item":                        "vcd_catalog_vapp_template",
	"vcd_org_ldap":                            "", // not needed: handled by vcd_org
	"vcd_vapp_vm":                             "", // not needed: handled by vcd_vapp
	"vcd_vapp_network":                        "", // not needed: handled by vcd_vapp
	"vcd_vapp_static_routing":                 "", // not needed: handled by vcd_vapp
	"vcd_vapp_firewall_rules":                 "", // not needed: handled by vcd_vapp
	"vcd_edgegateway":                         "", // not handled yet
	"vcd_external_network":                    "", // not handled yet
	"vcd_lb_app_profile":                      "", // not handled yet
	"vcd_lb_app_rule":                         "", // not handled yet
	"vcd_lb_server_pool":                      "", // not handled yet
	"vcd_lb_service_monitor":                  "", // not handled yet
	"vcd_nsxt_alb_cloud":                      "", // not handled yet
	"vcd_nsxt_alb_service_engine_group":       "", // not handled yet
	"vcd_nsxt_app_port_profile":               "", // not handled yet
	"vcd_nsxt_distributed_firewall":           "", // not handled yet
	"vcd_nsxt_edgegateway":                    "", // not handled yet
	"vcd_nsxt_edgegateway_bgp_ip_prefix_list": "", // not handled yet
	"vcd_nsxt_network_imported":               "", // not handled yet
	"vcd_nsxv_firewall_rule":                  "", // not handled yet
	"vcd_nsxv_ip_set":                         "", // not handled yet
	"vcd_org_user":                            "", // not handled yet
	"vcd_security_tag":                        "", // not handled yet
	"vcd_vm_placement_policy":                 "", // not handled yet
	"vcd_vm_sizing_policy":                    "", // not handled yet
}

func main() {
	// No arguments are needed. The program loops through the HCL files in directory ./vcd/test-artifacts
	fileList, err := filepath.Glob("vcd.*.tf")
	if err != nil {
		fmt.Printf("error reading directory %s\n", err)
		os.Exit(1)
	}

	// usedResources keeps trace of resources we have already seen
	var usedResources = make(map[string]bool)

	// foundEntities collects the resources type and name
	var foundEntities entityList

	//resourceRegexp is a regular expression that finds a resource text
	//  (?ms)        m = set multi-line mode  - s = let '.' match newlines
	//  ^resource    finds "resource" at the start of a line
	//  \s*          followed by optional spaces
	//  "(\w+)"      captures the text within quotes (resource type)
	//  \s*          followed by optional spaces
	//  "([^"]+)     captures the text within quotes (resource name in state)
	//  "\s*         followed by optional spaces
	//  (\{.+?^})    finds an opening '{', followed by any character, until a closing '}' at the start of a line
	//               The '?' limits the greediness of the expression, matching the first closing brace.
	//
	// Note: the regexp assumes that the HCL files are properly formatted, i.e. that there is no
	// closing '}' at the start of the line before the end of the resource
	resourceRegexp := regexp.MustCompile(`(?ms)^resource\s*"(\w+)"\s*"([^"]+)"\s*(\{.+?^})`)

	// nameRegexp retrieves the name of the resource within the resource text
	//  (?m)        set multi-line mode (^ matches the start of the line, not the whole string)
	//  ^\s*name   find "name" at the start of a line, with optional spaces before it
	//  \s*        followed by optional spaces
	//  =          the 'equals' sign
	//  \s*        followed by optional spaces
	//  "([^"]+)"  captures the text within quotes (the entity name)
	nameRegexp := regexp.MustCompile(`(?m)^\s*name\s*=\s*"([^"]+)"`)

	// countRegexp retrieves the resource count within the resource text
	//  (?m)        set multi-line mode (^ matches the start of the line, not the whole string)
	//  ^\s*count   find "count" at the start of a line, with optional spaces before it
	//  \s*         followed by optional spaces
	//  =           the 'equals' sign
	//  \s*         followed by optional spaces
	//  ([0-9]+)    captures the number (count)
	countRegexp := regexp.MustCompile(`(?m)^\s*count\s*=\s*([0-9]+)`)

	// resourceIndex is the progressive count of the resources found
	resourceIndex := 0

	for _, fileName := range fileList {

		// First, we get the file contents
		text, err := os.ReadFile(fileName) // #nosec G304 -- only used for test creation
		if err != nil {
			fmt.Printf("error reading %s: %s\n", fileName, err)
			os.Exit(1)
		}

		// Then, we collect each resource in the file
		matches := resourceRegexp.FindAllStringSubmatch(string(text), -1)

		// Every item is a string slice, capturing type, definitionName, and body of the resource
		for _, item := range matches {
			resourceType := item[1]
			resourceDef := item[2]
			resourceText := item[3]

			// within the resource text, we search for the entity name
			resourceNameMatches := nameRegexp.FindStringSubmatch(resourceText)
			resourceName := ""
			if len(resourceNameMatches) > 0 {
				resourceName = resourceNameMatches[1]
			}

			// Names that started with '[Tt]est' are skipped. They are picked automatically
			// by the leftover removal tool
			if strings.HasPrefix(resourceName, "Test") || strings.HasPrefix(resourceName, "test") {
				continue
			}

			// Get the ultimate resource type, i.e. the one we support in the leftovers removal tool
			newType, found := typeReplacements[resourceType]
			if found {
				resourceType = newType
			}

			// If there is no resource, we skip this resource, as it is not handled yet.
			if resourceType == "" {
				continue
			}

			// We try to avoid duplicates. A pair of type+name will be enough, regardless of their parent resource
			id := resourceType + "#" + resourceName
			_, seen := usedResources[id]
			if seen {
				continue
			}
			// Set the used resource record, to avoid finding the item if it shows up again
			usedResources[id] = true

			// When there is no name, it's because the resource can only be identified by ID
			if resourceName == "" {
				continue
			}

			// When the resource name contains a count, we need to find the count and unroll it
			if strings.Contains(resourceName, "${count.index}") {
				countMatches := countRegexp.FindStringSubmatch(resourceText)
				if len(countMatches) > 1 {
					countText := countMatches[1]
					count, err := strconv.Atoi(countText)
					if err != nil {
						fmt.Printf("#file: %s - resource %s.%s (%s): %s\n", fileName, resourceType, resourceDef, resourceName, err)
						continue
					}
					// Once we have found the count, we create as many names as the count requires
					for c := 0; c < count; c++ {
						resourceIndex++
						name := strings.Replace(resourceName, "${count.index}", fmt.Sprintf("%d", c), 1)

						id := resourceType + "#" + name
						_, seen := usedResources[id]
						if seen {
							continue
						}
						usedResources[id] = true
						foundEntities = append(foundEntities, entityDef{
							Type:    resourceType,
							Name:    name,
							Comment: fmt.Sprintf("%d - from %s: %s (%s count = %d)", resourceIndex, fileName, resourceDef, resourceName, count),
						})
					}

				}
			} else {
				// There was no count in the name. We just handle it
				resourceIndex++
				foundEntities = append(foundEntities, entityDef{
					Type:    resourceType,
					Name:    resourceName,
					Comment: fmt.Sprintf("%d - from %s: %s", resourceIndex, fileName, resourceName),
				})
			}
		}
	}

	jsonText, err := json.MarshalIndent(foundEntities, " ", " ")
	if err != nil {
		fmt.Printf("error encoding JSON: %s\n", err)
		os.Exit(1)
	}

	// At the end we print two things:
	// 1: the JSON text of our entities (to be used stand-alone)
	fmt.Printf("%s\n", jsonText)
	// 2: the Go representation of the entities (to be pasted in remove_leftovers_test.go)
	fmt.Printf("%# v\n", pretty.Formatter(foundEntities))
}
