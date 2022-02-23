package handlers

import (
	"fmt"
	"regexp"
	"strconv"
)

func getItemID(regex *regexp.Regexp, exp string) (uint64, error) {
	// parse id from expression
	matches := regex.FindStringSubmatch(exp)
	if len(matches) < 2 {
		return 0, fmt.Errorf("id not found")
	}

	// convert id to integer
	id, _ := strconv.Atoi(matches[1])

	return uint64(id), nil
}

func getQueryParams(query string) (limit int, sort string) {
	limit = 0
	sort = "asc"

	queryRegex := regexp.MustCompile(`(limit=(\d+))|(sort=(asc|desc))`)
	matches := queryRegex.FindStringSubmatch(query)
	if matches != nil {
		if matches[2] != "" {
			limit, _ = strconv.Atoi(matches[2])
		} else if matches[4] != "" {
			sort = matches[4]
		}
	}

	return
}
