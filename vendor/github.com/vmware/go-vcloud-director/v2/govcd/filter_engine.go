package govcd

import (
	"fmt"
	"regexp"
	"strings"
)

// SearchByFilter is a generic filter that can operate on entities that implement the QueryItem interface
// It requires a queryType and a set of criteria.
// Returns a list of QueryItem interface elements, which can be cast back to the wanted real type
func (client *Client) SearchByFilter(queryType string, criteria *FilterDef) ([]QueryItem, error) {

	// Set of conditions to be evaluated
	var conditions []conditionDef
	// List of candidate items that match all conditions
	var candidatesByConditions []QueryItem
	// item with the latest date among the candidates
	var candidateByLatest QueryItem

	// By setting the latest date to the early possible date, we make sure that it will be swapped
	// at the first comparison
	var latestDate = "1970-01-01 00:00:00"

	var metadataFields []string
	// If set, metadata fields will be passed as 'metadata@SYSTEM:fieldName'
	var isSystem bool

	// Will search the latest item if requested
	searchLatest := false

	// Parse criteria and build the condition list
	for key, value := range criteria.Filters {
		switch key {
		case FilterNameRegex:
			if value != "" {
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling regular expression '%s' : %s ", value, err)
				}
				conditions = append(conditions, conditionDef{key, nameCondition{re}})
			}
		case FilterDate:
			if value != "" {
				conditions = append(conditions, conditionDef{key, dateCondition{value}})
			}
		case FilterIp:
			if value != "" {
				re, err := regexp.Compile(value)
				if err != nil {
					return nil, fmt.Errorf("error compiling regular expression '%s' : %s ", value, err)
				}
				conditions = append(conditions, conditionDef{key, ipCondition{re}})
			}
		case FilterLatest:
			searchLatest = StringToBool(value)

		default:
			return nil, fmt.Errorf("[SearchByFilter] filter '%s' not supported (only allowed %v), %s)", key, value, SupportedFilters)
		}
	}

	var metadataFilter = make(map[string]MetadataFilter)
	if len(criteria.Metadata) > 0 {
		for _, cond := range criteria.Metadata {
			k := cond.Key
			v := cond.Value
			isSystem = cond.IsSystem
			if criteria.UseMetadataApiFilter && cond.Type != "" && !strings.EqualFold(cond.Type, "none") {
				metadataFilter[cond.Key] = MetadataFilter{
					Type:  cond.Type,
					Value: fmt.Sprintf("%v", cond.Value),
				}
			}
			metadataFields = append(metadataFields, k)
			re, err := regexp.Compile(v.(string))
			if err != nil {
				return nil, fmt.Errorf("error compiling regular expression '%s' : %s ", v, err)
			}
			if !criteria.UseMetadataApiFilter {
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
		itemResult, err = client.QueryByMetadataFilter(nil, params, metadataFilter, isSystem)
	} else {
		itemResult, err = client.QueryWithMetadataFields(queryType, nil, params, metadataFields, isSystem)
	}

	if err != nil {
		return nil, fmt.Errorf("[SearchByFilter] error retrieving query item list: %s", err)
	}
	var itemList []QueryItem

	itemList, err = resultsToQueryItem(queryType, itemResult)
	if err != nil {
		return nil, fmt.Errorf("[SearchByFilter] error converting QueryItem  item list: %s", err)
	}

	for _, item := range itemList {
		numOfMatches := 0

		for _, condition := range conditions {
			result, err := conditionMatches(condition.conditionType, condition.stored, item)
			if err != nil {
				return nil, fmt.Errorf("[SearchByFilter] error applying condition %v: %s", condition, err)
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
		return nil, fmt.Errorf("[SearchByFilter] no items found with given criteria %s", conditionText(criteria))
	}
	if searchLatest {
		for _, candidate := range candidatesByConditions {
			greater, err := compareDate(fmt.Sprintf("> %s", latestDate), candidate.GetDate())
			if err != nil {
				return nil, fmt.Errorf("[SearchByFilter] error comparing dates %s > %s",
					candidate.GetDate(), latestDate)
			}
			if greater {
				latestDate = candidate.GetDate()
				candidateByLatest = candidate.(QueryItem)
			}
		}
		if candidateByLatest != nil {
			return []QueryItem{candidateByLatest}, nil
		}
	}
	return candidatesByConditions, nil
}

// conditionMatches performs the appropriate condition evaluation,
// depending on conditionType
func conditionMatches(conditionType string, stored, item interface{}) (bool, error) {
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
	return false, fmt.Errorf("unsupported condition type '%s'", conditionType)
}
