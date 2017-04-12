package core

import (
	"strings"
)

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
