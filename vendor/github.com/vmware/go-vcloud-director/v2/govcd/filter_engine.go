package govcd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kr/pretty"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type queryWithMetadataFunc func(queryType string, params, notEncodedParams map[string]string,
	metadataFields []string, isSystem bool) (Results, error)

type queryByMetadataFunc func(queryType string, params, notEncodedParams map[string]string,
	metadataFilters map[string]MetadataFilter, isSystem bool) (Results, error)

type resultsConverterFunc func(queryType string, results Results) ([]QueryItem, error)

// searchByFilter is a generic filter that can operate on entities that implement the QueryItem interface
// It requires a queryType and a set of criteria.
// Returns a list of QueryItem interface elements, which can be cast back to the wanted real type
// Also returns a human readable text of the conditions being passed and how they matched the data found
func searchByFilter(queryByMetadata queryByMetadataFunc, queryWithMetadataFields queryWithMetadataFunc,
	converter resultsConverterFunc, queryType string, criteria *FilterDef) ([]QueryItem, string, error) {

	// Set of conditions to be evaluated (will be filled from criteria)
	var conditions []conditionDef
	// List of candidate items that match all conditions
	var candidatesByConditions []QueryItem

	// List of metadata fields that will be added to the query
	var metadataFields []string

	// If set, metadata fields will be passed as 'metadata@SYSTEM:fieldName'
	var isSystem bool
	var params = make(map[string]string)

	// Will search the latest item if requested
	searchLatest := false
	// Will search the earliest item if requested
	searchEarliest := false

	// A null filter is converted into an empty object.
	// Using an empty filter is equivalent to fetching all items without filtering
	if criteria == nil {
		criteria = &FilterDef{}
	}

	// A text containing the human-readable form of the criteria being used, and the detail on how they matched the
	// data being fetched
	explanation := conditionText(criteria)

	// A collection of matching information for the conditions being applied
	var matches []matchResult

	// Parse criteria and build the condition list
	for key, value := range criteria.Filters {
		// Empty values could be leftovers from the criteria build-up prior to calling this function
		if value == "" {
			continue
		}
		switch key {
		case types.FilterNameRegex:
			re, err := regexp.Compile(value)
			if err != nil {
				return nil, explanation, fmt.Errorf("error compiling regular expression '%s' : %s ", value, err)
			}
			conditions = append(conditions, conditionDef{key, nameCondition{re}})
		case types.FilterDate:
			conditions = append(conditions, conditionDef{key, dateCondition{value}})
		case types.FilterIp:
			re, err := regexp.Compile(value)
			if err != nil {
				return nil, explanation, fmt.Errorf("error compiling regular expression '%s' : %s ", value, err)
			}
			conditions = append(conditions, conditionDef{key, ipCondition{re}})
		case types.FilterParent:
			conditions = append(conditions, conditionDef{key, parentCondition{value}})
		case types.FilterParentId:
			conditions = append(conditions, conditionDef{key, parentIdCondition{value}})

		case types.FilterLatest:
			searchLatest = stringToBool(value)

		case types.FilterEarliest:
			searchEarliest = stringToBool(value)

		default:
			return nil, explanation, fmt.Errorf("[SearchByFilter] filter '%s' not supported (only allowed %v)", key, supportedFilters)
		}
	}

	// We can't allow the search for both the oldest and the newest item
	if searchEarliest && searchLatest {
		return nil, explanation, fmt.Errorf("only one of '%s' or '%s' can be used for a set of criteria", types.FilterEarliest, types.FilterLatest)
	}

	var metadataFilter = make(map[string]MetadataFilter)
	// Fill metadata filters
	if len(criteria.Metadata) > 0 {
		for _, cond := range criteria.Metadata {
			k := cond.Key
			v := cond.Value
			isSystem = cond.IsSystem
			if k == "" {
				return nil, explanation, fmt.Errorf("metadata condition without key detected")
			}
			if v == "" {
				return nil, explanation, fmt.Errorf("empty value for metadata condition with key '%s'", k)
			}

			// If we use the metadata search through the API, we must make sure that the type is set
			if criteria.UseMetadataApiFilter {
				if cond.Type == "" || strings.EqualFold(cond.Type, "none") {
					return nil, explanation, fmt.Errorf("requested search by metadata field '%s' must provide a valid type", cond.Key)
				}

				// The type must be one of the expected values
				err := validateMetadataType(cond.Type)
				if err != nil {
					return nil, explanation, fmt.Errorf("type '%s' for metadata field '%s' is invalid. :%s", cond.Type, cond.Key, err)
				}
				metadataFilter[cond.Key] = MetadataFilter{
					Type:  cond.Type,
					Value: fmt.Sprintf("%v", cond.Value),
				}
			}

			// If we don't use metadata search via the API, we add the field to the list, and
			// also add a condition, using regular expressions
			if !criteria.UseMetadataApiFilter {
				metadataFields = append(metadataFields, k)
				re, err := regexp.Compile(v.(string))
				if err != nil {
					return nil, explanation, fmt.Errorf("error compiling regular expression '%s' : %s ", v, err)
				}
				conditions = append(conditions, conditionDef{"metadata", metadataRegexpCondition{k, re}})
			}
		}
	} else {
		criteria.UseMetadataApiFilter = false
	}

	var itemResult Results
	var err error

	if criteria.UseMetadataApiFilter {
		// This result will not include metadata fields. The query will use metadata parameters to restrict the search
		itemResult, err = queryByMetadata(queryType, nil, params, metadataFilter, isSystem)
	} else {
		// This result includes metadata fields, if they exist.
		itemResult, err = queryWithMetadataFields(queryType, nil, params, metadataFields, isSystem)
	}

	if err != nil {
		return nil, explanation, fmt.Errorf("[SearchByFilter] error retrieving query item list: %s", err)
	}
	if dataInspectionRequested("QE1") {
		util.Logger.Printf("[INSPECT-QE1-SearchByFilter] list of retrieved items %# v\n", pretty.Formatter(itemResult.Results))
	}
	var itemList []QueryItem

	// Converting the query result into a list of QueryItems
	itemList, err = converter(queryType, itemResult)
	if err != nil {
		return nil, explanation, fmt.Errorf("[SearchByFilter] error converting QueryItem  item list: %s", err)
	}
	if dataInspectionRequested("QE2") {
		util.Logger.Printf("[INSPECT-QE2-SearchByFilter] list of converted items %# v\n", pretty.Formatter(itemList))
	}

	// Process the list using the conditions gathered above
	for _, item := range itemList {
		numOfMatches := 0

		for _, condition := range conditions {

			if dataInspectionRequested("QE3") {
				util.Logger.Printf("[INSPECT-QE3-SearchByFilter]\ncondition %# v\nitem %# v\n", pretty.Formatter(condition), pretty.Formatter(item))
			}
			result, definition, err := conditionMatches(condition.conditionType, condition.stored, item)
			if err != nil {
				return nil, explanation, fmt.Errorf("[SearchByFilter] error applying condition %v: %s", condition, err)
			}

			// Saves matching information, which will be consolidated in the final explanation text
			matches = append(matches, matchResult{
				Name:       item.GetName(),
				Type:       condition.conditionType,
				Definition: definition,
				Result:     result,
			})
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

	// Consolidates the explanation with information about which conditions did actually match
	matchesText := matchesToText(matches)
	explanation += fmt.Sprintf("\n%s", matchesText)
	util.Logger.Printf("[SearchByFilter] conditions matching\n%s", explanation)

	// Once all the conditions have been evaluated, we check whether we got any items left.
	//
	// We consider an empty result to be a valid one: it's up to the caller to evaluate the result
	// and eventually use the explanation to provide an error message
	if len(candidatesByConditions) == 0 {
		return nil, explanation, nil
	}

	// If we have only one item, there is no reason to search further for the newest or oldest item
	if len(candidatesByConditions) == 1 {
		return candidatesByConditions, explanation, nil
	}
	var emptyDatesFound []string
	if searchLatest {
		// By setting the latest date to the early possible date, we make sure that it will be swapped
		// at the first comparison
		var latestDate = "1970-01-01 00:00:00"
		// item with the latest date among the candidates
		var candidateByLatest QueryItem
		for _, candidate := range candidatesByConditions {
			itemDate := candidate.GetDate()
			if itemDate == "" {
				emptyDatesFound = append(emptyDatesFound, candidate.GetName())
				continue
			}
			util.Logger.Printf("[SearchByFilter] search latest: comparing %s to %s", latestDate, itemDate)
			greater, err := compareDate(fmt.Sprintf("> %s", latestDate), itemDate)
			if err != nil {
				return nil, explanation, fmt.Errorf("[SearchByFilter] error comparing dates %s > %s : %s",
					candidate.GetDate(), latestDate, err)
			}
			util.Logger.Printf("[SearchByFilter] result %v: ", greater)
			if greater {
				latestDate = candidate.GetDate()
				candidateByLatest = candidate.(QueryItem)
			}
		}
		if candidateByLatest != nil {
			explanation += "\nlatest item found"
			return []QueryItem{candidateByLatest}, explanation, nil
		} else {
			return nil, explanation, fmt.Errorf("search for newest item failed. Empty dates found for items %v", emptyDatesFound)
		}
	}
	if searchEarliest {
		// earliest date is set to a date in the future (10 years from now), so that any date found will be evaluated as
		// earlier than this one
		var earliestDate = time.Now().AddDate(10, 0, 0).String()
		// item with the earliest date among the candidates
		var candidateByEarliest QueryItem
		for _, candidate := range candidatesByConditions {
			itemDate := candidate.GetDate()
			if itemDate == "" {
				emptyDatesFound = append(emptyDatesFound, candidate.GetName())
				continue
			}
			util.Logger.Printf("[SearchByFilter] search earliest: comparing %s to %s", earliestDate, candidate.GetDate())
			greater, err := compareDate(fmt.Sprintf("< %s", earliestDate), candidate.GetDate())
			if err != nil {
				return nil, explanation, fmt.Errorf("[SearchByFilter] error comparing dates %s > %s: %s",
					candidate.GetDate(), earliestDate, err)
			}
			util.Logger.Printf("[SearchByFilter] result %v: ", greater)
			if greater {
				earliestDate = candidate.GetDate()
				candidateByEarliest = candidate.(QueryItem)
			}
		}
		if candidateByEarliest != nil {
			explanation += "\nearliest item found"
			return []QueryItem{candidateByEarliest}, explanation, nil
		} else {
			return nil, explanation, fmt.Errorf("search for oldest item failed. Empty dates found for items %v", emptyDatesFound)
		}
	}
	if searchEarliest || searchLatest {
		// We should never reach this point, as a failure for newest or oldest item was caught above, but just in case
		return nil, explanation, fmt.Errorf("search for oldest or earliest item failed. No reason found")
	}
	return candidatesByConditions, explanation, nil
}

// conditionMatches performs the appropriate condition evaluation,
// depending on conditionType
func conditionMatches(conditionType string, stored, item interface{}) (bool, string, error) {
	switch conditionType {
	case types.FilterNameRegex:
		return matchName(stored, item)
	case types.FilterDate:
		return matchDate(stored, item)
	case types.FilterIp:
		return matchIp(stored, item)
	case types.FilterParent:
		return matchParent(stored, item)
	case types.FilterParentId:
		return matchParentId(stored, item)
	case "metadata":
		return matchMetadata(stored, item)
	}
	return false, "", fmt.Errorf("unsupported condition type '%s'", conditionType)
}

// SearchByFilter is a generic filter that can operate on entities that implement the QueryItem interface
// It requires a queryType and a set of criteria.
// Returns a list of QueryItem interface elements, which can be cast back to the wanted real type
// Also returns a human readable text of the conditions being passed and how they matched the data found
// See "## Query engine" in CODING_GUIDELINES.md for more info
func (client *Client) SearchByFilter(queryType string, criteria *FilterDef) ([]QueryItem, string, error) {
	return searchByFilter(client.queryByMetadataFilter, client.queryWithMetadataFields, resultToQueryItems, queryType, criteria)
}

// SearchByFilter runs the search for a specific catalog
// The 'parentField' argument defines which filter will be added, depending on the items we search for:
//   - 'catalog' contains the catalog HREF or ID
//   - 'catalogName' contains the catalog name
func (catalog *Catalog) SearchByFilter(queryType, parentField string, criteria *FilterDef) ([]QueryItem, string, error) {
	var err error
	switch parentField {
	case "catalog":
		err = criteria.AddFilter(types.FilterParentId, catalog.Catalog.ID)
	case "catalogName":
		err = criteria.AddFilter(types.FilterParent, catalog.Catalog.Name)
	default:
		return nil, "", fmt.Errorf("unrecognized filter field '%s'", parentField)
	}
	if err != nil {
		return nil, "", fmt.Errorf("error setting parent filter for catalog %s with fieldName '%s'", catalog.Catalog.Name, parentField)
	}
	return catalog.client.SearchByFilter(queryType, criteria)
}

// SearchByFilter runs the search for a specific VDC
// The 'parentField' argument defines which filter will be added, depending on the items we search for:
//   - 'vdc' contains the VDC HREF or ID
//   - 'vdcName' contains the VDC name
func (vdc *Vdc) SearchByFilter(queryType, parentField string, criteria *FilterDef) ([]QueryItem, string, error) {
	var err error
	switch parentField {
	case "vdc":
		err = criteria.AddFilter(types.FilterParentId, vdc.Vdc.ID)
	case "vdcName":
		err = criteria.AddFilter(types.FilterParent, vdc.Vdc.Name)
	default:
		return nil, "", fmt.Errorf("unrecognized filter field '%s'", parentField)
	}
	if err != nil {
		return nil, "", fmt.Errorf("error setting parent filter for VDC %s with fieldName '%s'", vdc.Vdc.Name, parentField)
	}
	return vdc.client.SearchByFilter(queryType, criteria)
}

// SearchByFilter runs the search for a specific Org
func (org *AdminOrg) SearchByFilter(queryType string, criteria *FilterDef) ([]QueryItem, string, error) {
	err := criteria.AddFilter(types.FilterParent, org.AdminOrg.Name)
	if err != nil {
		return nil, "", fmt.Errorf("error setting parent filter for Org %s with fieldName 'orgName'", org.AdminOrg.Name)
	}
	return org.client.SearchByFilter(queryType, criteria)
}

// SearchByFilter runs the search for a specific Org
func (org *Org) SearchByFilter(queryType string, criteria *FilterDef) ([]QueryItem, string, error) {
	err := criteria.AddFilter(types.FilterParent, org.Org.Name)
	if err != nil {
		return nil, "", fmt.Errorf("error setting parent filter for Org %s with fieldName 'orgName'", org.Org.Name)
	}
	return org.client.SearchByFilter(queryType, criteria)
}

// dataInspectionRequested checks if the given code was found in the inspection environment variable.
func dataInspectionRequested(code string) bool {
	govcdInspect := os.Getenv("GOVCD_INSPECT")
	return strings.Contains(govcdInspect, code)
}
