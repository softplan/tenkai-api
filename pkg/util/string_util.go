package util

import (
	"strings"
)

//GetReplacebleKeyName - Retrieve variable names
func GetReplacebleKeyName(value string) []string {
	var result []string
	for {
		s := strings.Index(value, "${")
		e := strings.Index(value, "}")
		r := getVariableValue(value, s, e)
		if r != "" {
			result = append(result, r)
		} else {
			break
		}
		value = value[e+1:]
	}
	return result
}

func getVariableValue(value string, start int, end int) string {
	if start > -1 {
		if (end > -1) && (end > start) {
			return value[start+2 : end]
		}
	}
	return ""
}

//Contains Contains
func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}
