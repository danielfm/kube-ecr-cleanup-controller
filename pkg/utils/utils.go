package utils

import (
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecr"
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

// ApplyKeepFilters takes a list of images and removes those matching the
// filters.
func ApplyKeepFilters(images []*ecr.ImageDetail, filters []*string) []*ecr.ImageDetail {
	filtered := make([]*ecr.ImageDetail, 0)
	regs := []*regexp.Regexp{}
	for _, filter := range filters {
		reg, _ := regexp.Compile(*filter)
		regs = append(regs, reg)
	}

	for _, image := range images {
		keep := false
		for _, tag := range image.ImageTags {
			for _, reg := range regs {
				if reg.MatchString(*tag) {
					keep = true
				}
			}
		}
		if !keep {
			filtered = append(filtered, image)
		}
	}

	return filtered
}
