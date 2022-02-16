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
