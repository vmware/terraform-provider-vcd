package vcd

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// --------------------------------------------------------------------------
// vApp template search interface
// --------------------------------------------------------------------------

type ConditionDate struct {
	dateExpression string
}

type ConditionRegexp struct {
	regExpression *regexp.Regexp
}

// VappTemplateConditionDate supports search by date
type VappTemplateConditionDate ConditionDate

// VappTemplateConditionRegexp supports search by name regex
type VappTemplateConditionRegexp ConditionRegexp

type VappTemplateMetadataConditionRegexp struct {
	fieldName     string
	regExpression *regexp.Regexp
}

// VappTemplateCondition is an interface that allows different conditions to be
// used with the same calls
type VappTemplateCondition interface {
	matches(item *types.QueryResultVappTemplateType) (bool, error)
	value(item *types.QueryResultVappTemplateType) string
}

// VappTemplateConditionDate.value returns the value being used for date comparison
func (cd VappTemplateConditionDate) value(item *types.QueryResultVappTemplateType) string {
	return item.CreationDate
}

// VappTemplateConditionDate.matches evaluates the expression against the internal value
func (cd VappTemplateConditionDate) matches(item *types.QueryResultVappTemplateType) (bool, error) {
	return compareDate(cd.dateExpression, cd.value(item))
}

// VappTemplateConditionRegexp.value returns the value being used for regex comparison
func (cr VappTemplateConditionRegexp) value(item *types.QueryResultVappTemplateType) string {
	return item.Name
}

// VappTemplateConditionRegexp.matches evaluates the regular expression against the item name
func (cr VappTemplateConditionRegexp) matches(item *types.QueryResultVappTemplateType) (bool, error) {
	return cr.regExpression.MatchString(cr.value(item)), nil
}

func (cr VappTemplateMetadataConditionRegexp) value(item *types.QueryResultVappTemplateType) string {
	if item.Metadata == nil || len(item.Metadata.MetadataEntry) == 0 {
		return ""
	}
	for _, x := range item.Metadata.MetadataEntry {
		if cr.fieldName == x.Key {
			return x.TypedValue.Value
		}
	}
	return ""
}

func (cr VappTemplateMetadataConditionRegexp) matches(item *types.QueryResultVappTemplateType) (bool, error) {
	value := cr.value(item)
	if value == "" {
		return false, nil
	}
	return cr.regExpression.MatchString(value), nil
}

// vappTemplateToCatalogItem returns the catalog item corresponding to the associated vApp template
func vappTemplateToCatalogItem(vappTemplate *types.QueryResultVappTemplateType, catalog *govcd.Catalog) (*govcd.CatalogItem, error) {
	item, err := catalog.GetCatalogItemByName(vappTemplate.Name, false)
	if err != nil {
		return nil, fmt.Errorf("error retrieving catalog item from vapp template %s: %s", vappTemplate.Name, err)
	}
	return item, nil
}

// searchVappTemplateByFilter searches an item within a list of catalog items
// - catalog is the items container
// - filter is a map of values containing the expressions to be evaluated
//   Currently supported:
//     * name_regex: a regular expression for the name
//     * date: a combination of an operator (>, >=, ==, <, <=) and a date
//     * latest: if set, the most recent item is selected
//     * metadata: a set of conditions made of 'key', 'value', and 'is_system'
func searchVappTemplateByFilter(catalog *govcd.Catalog, filter interface{}, isSysAdmin bool) (*govcd.CatalogItem, error) {

	// Note: we search a list of vApp templates because the metadata, if present, is attached
	// to the vApp template, not the Catalog Item.
	// When the vApp template is found, the corresponding catalog item is returned through vappTemplateToCatalogItem

	// Set of conditions to be evaluated
	var conditions []VappTemplateCondition
	// List of candidate items that match all conditions
	var candidatesByConditions []*types.QueryResultVappTemplateType
	// item with the latest date among the candidates
	var candidateByLatest *types.QueryResultVappTemplateType

	// By setting the latest date to the early possible date, we make sure that it will be swapped
	// at the first comparison
	var latestDate = "1970-01-01 00:00:00"

	var metadataFields []string
	// If set, metadata fields will be passed as 'metadata@SYSTEM:fieldName'
	var isSystem bool

	// Will search the latest item if requested
	searchLatest := false

	filters, ok := filter.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter parameter must be a slice of interface{}. (Currently %#v)", filter)
	}
	if len(filters) == 0 {
		return nil, fmt.Errorf("empty criteria")
	}
	criteria, ok := filters[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("filter parameter must be a map of values. (Currently %#v)", criteria)
	}

	// Parse criteria and build the condition list
	for key, value := range criteria {
		switch key {
		case "name_regex":
			if value.(string) != "" {
				conditions = append(conditions, VappTemplateConditionRegexp{regexp.MustCompile(value.(string))})
			}

		case "date":
			if value.(string) != "" {
				conditions = append(conditions, VappTemplateConditionDate{value.(string)})
			}
		case "latest":
			searchLatest = value.(bool)
		case "metadata":
			metadataConditions := value.([]interface{})
			for _, mdc := range metadataConditions {
				condition := mdc.(map[string]interface{})
				k := condition["key"].(string)
				v := condition["value"].(string)
				isSystem = condition["is_system"].(bool)
				metadataFields = append(metadataFields, k)
				conditions = append(conditions, VappTemplateMetadataConditionRegexp{k, regexp.MustCompile(v)})
			}

		default:
			return nil, fmt.Errorf("[searchVappTemplateByFilter] filter '%s' not supported (only 'name_regex', 'date', 'latest', and 'metadata' allowed, %s)", key, value)
		}
	}

	queryType := "vappTemplate"

	if isSysAdmin {
		queryType = "adminVAppTemplate"
	}
	params := map[string]string{
		"filter": fmt.Sprintf("catalogName==%s", url.QueryEscape(catalog.Catalog.Name)),
	}
	itemResult, err := catalog.QueryWithMetadataFields(queryType, nil, params, metadataFields, isSystem)
	if err != nil {
		return nil, fmt.Errorf("[searchVappTemplateByFilter] error retrieving catalog item list: %s", err)
	}
	var itemList []*types.QueryResultVappTemplateType
	if isSysAdmin {
		itemList = itemResult.Results.AdminVappTemplateRecord
	} else {
		itemList = itemResult.Results.VappTemplateRecord
	}

	for _, item := range itemList {
		numOfMatches := 0

		for _, condition := range conditions {
			result, err := condition.matches(item)
			if err != nil {
				return nil, fmt.Errorf("[searchVappTemplateByFilter] error applying condition %v: %s", condition, err)
			}
			if !result {
				continue
			}
			numOfMatches++
		}
		if numOfMatches == len(conditions) {
			// All conditions were met
			candidatesByConditions = append(candidatesByConditions, item)
		}
	}
	if len(candidatesByConditions) == 0 {
		return nil, fmt.Errorf("[searchVappTemplateByFilter] no items found with given criteria")
	}
	if searchLatest {
		for _, vappTemplate := range candidatesByConditions {
			greater, err := compareDate(fmt.Sprintf("> %s", latestDate), vappTemplate.CreationDate)
			if err != nil {
				return nil, fmt.Errorf("[searchVappTemplateByFilter] error comparing dates %s > %s",
					vappTemplate.CreationDate, latestDate)
			}
			if greater {
				latestDate = vappTemplate.CreationDate
				candidateByLatest = vappTemplate
			}
		}
		if candidateByLatest != nil {
			return vappTemplateToCatalogItem(candidateByLatest, catalog)
		}
	}
	if len(candidatesByConditions) > 1 {
		var itemNames = make([]string, len(candidatesByConditions))
		for i, item := range candidatesByConditions {
			itemNames[i] = item.Name
		}
		return nil, fmt.Errorf("found multiple items with given criteria: %v", itemNames)
	}
	// At this point, there is only one item remaining
	return vappTemplateToCatalogItem(candidatesByConditions[0], catalog)
}
