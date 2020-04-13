package govcd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/util"
)

type queryWithMetadataFunc func(queryType string, params, notEncodedParams map[string]string,
	metadataFields []string, isSystem bool) (Results, error)

type queryByMetadataFunc func(params, notEncodedParams map[string]string,
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
	// item with the latest date among the candidates
	var candidateByLatest QueryItem
	var candidateByEarliest QueryItem

	// By setting the latest date to the early possible date, we make sure that it will be swapped
	// at the first comparison
	var latestDate = "0001-01-01 00:00:00"

	// earliest date is set to a date in the future (100 years from now), so that any date found will be evaluated as
	// earlier than this one
	var earliestDate = time.Now().AddDate(100, 0, 0).String()

	// List of metadata fields that will be added to the query
	var metadataFields []string

	// If set, metadata fields will be passed as 'metadata@SYSTEM:fieldName'
	var isSystem bool

	// Will search the latest item if requested
	searchLatest := false
	// Will search the earliest item if requested
	searchEarliest := false

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
		case FilterNameRegex:
			re, err := regexp.Compile(value)
			if err != nil {
				return nil, explanation, fmt.Errorf("error compiling regular expression '%s' : %s ", value, err)
			}
			conditions = append(conditions, conditionDef{key, nameCondition{re}})
		case FilterDate:
			conditions = append(conditions, conditionDef{key, dateCondition{value}})
		case FilterIp:
			re, err := regexp.Compile(value)
			if err != nil {
				return nil, explanation, fmt.Errorf("error compiling regular expression '%s' : %s ", value, err)
			}
			conditions = append(conditions, conditionDef{key, ipCondition{re}})
		case FilterLatest:
			searchLatest = stringToBool(value)

		case FilterEarliest:
			searchEarliest = stringToBool(value)

		default:
			return nil, explanation, fmt.Errorf("[SearchByFilter] filter '%s' not supported (only allowed %v), %s)", key, value, SupportedFilters)
		}
	}

	// We can't allow the search for both the oldest and the newest item
	if searchEarliest && searchLatest {
		return nil, explanation, fmt.Errorf("only one of '%s' or '%s' can be used for a set of criteria", FilterEarliest, FilterLatest)
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
	params := map[string]string{}

	if criteria.UseMetadataApiFilter {
		params["type"] = queryType
		// This result will not include metadata fields. The query will use metadata parameters to restrict the search
		itemResult, err = queryByMetadata(nil, params, metadataFilter, isSystem)
	} else {
		// This result includes metadata fields, if they exist.
		itemResult, err = queryWithMetadataFields(queryType, nil, params, metadataFields, isSystem)
	}

	if err != nil {
		return nil, explanation, fmt.Errorf("[SearchByFilter] error retrieving query item list: %s", err)
	}
	var itemList []QueryItem

	// Converting the query result into a list of QueryItems
	itemList, err = converter(queryType, itemResult)
	if err != nil {
		return nil, explanation, fmt.Errorf("[SearchByFilter] error converting QueryItem  item list: %s", err)
	}

	// Process the list using the conditions gathered above
	for _, item := range itemList {
		numOfMatches := 0

		for _, condition := range conditions {
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
	case FilterNameRegex:
		return matchName(stored, item)
	case FilterDate:
		return matchDate(stored, item)
	case FilterIp:
		return matchIp(stored, item)
	case "metadata":
		return matchMetadata(stored, item)
	}
	return false, "", fmt.Errorf("unsupported condition type '%s'", conditionType)
}

// SearchByFilter is a generic filter that can operate on entities that implement the QueryItem interface
// It requires a queryType and a set of criteria.
// Returns a list of QueryItem interface elements, which can be cast back to the wanted real type
// Also returns a human readable text of the conditions being passed and how they matched the data found
func (client *Client) SearchByFilter(queryType string, criteria *FilterDef) ([]QueryItem, string, error) {
	return searchByFilter(client.QueryByMetadataFilter, client.QueryWithMetadataFields, resultToQueryItems, queryType, criteria)
}
