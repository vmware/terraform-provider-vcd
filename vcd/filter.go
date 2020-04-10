package vcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func buildMetadataCriteria(filter interface{}) ([]govcd.MetadataDef, bool, error) {
	var definitions []govcd.MetadataDef
	var useApiSearch bool
	filterList, ok := filter.([]interface{})
	if !ok {
		return nil, useApiSearch, fmt.Errorf("metadata block is not a list")
	}
	for _, raw := range filterList {
		metadataMap, ok := raw.(map[string]interface{})
		if !ok {
			return nil, useApiSearch, fmt.Errorf("metadata internal block is not a map")
		}
		var def govcd.MetadataDef
		for key, value := range metadataMap {
			switch key {
			case "key":
				def.Key = value.(string)
			case "value":
				def.Value = value
			case "type":
				def.Type = value.(string)
			case "is_system":
				def.IsSystem = value.(bool)
			case "use_api_search":
				useApiSearch = value.(bool)
			}
		}
		definitions = append(definitions, def)
	}

	return definitions, useApiSearch, nil
}

func buildCriteria(filter interface{}) (*govcd.FilterDef, error) {
	var criteria = govcd.NewFilterDef()

	filterList, ok := filter.([]interface{})
	if !ok {
		return nil, fmt.Errorf("[buildCriteria] filter is not a list")
	}

	filterMap, ok := filterList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("[buildCriteria] filter is not a map")
	}
	for key, value := range filterMap {
		switch key {

		case govcd.FilterNameRegex, govcd.FilterIp, govcd.FilterDate:
			err := criteria.AddFilter(key, value.(string))
			if err != nil {
				return nil, fmt.Errorf("[buildCriteria] error adding filter %s", err)
			}
		case govcd.FilterLatest:
			strValue := fmt.Sprintf("%v", value.(bool))
			err := criteria.AddFilter(key, strValue)
			if err != nil {
				return nil, fmt.Errorf("[buildCriteria] error adding filter %s", err)
			}
		case "metadata":
			definitions, useApiSearch, err := buildMetadataCriteria(value)
			if err != nil {
				return nil, fmt.Errorf("[buildCriteria] error adding metadata criteria %s", err)
			}
			criteria.UseMetadataApiFilter = useApiSearch
			criteria.Metadata = definitions
		}
	}
	return criteria, nil
}
