package core

import (
	"strings"
)

// ParseCommaSeparatedList takes a comma-separated string, such as "str1, str2",
// and returns a list of pointers to each element.
func ParseCommaSeparatedList(commaSeparatedList string) []*string {
	items := []*string{}
	rawList := strings.Split(commaSeparatedList, ",")

	for _, item := range rawList {
		trimmedItem := strings.TrimSpace(item)
		if len(trimmedItem) > 0 {
			items = append(items, &trimmedItem)
		}
	}

	return items
}
