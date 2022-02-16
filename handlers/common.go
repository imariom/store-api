package handlers

import (
	"fmt"
	"regexp"
	"strconv"
)

func getID(regex regexp.Regexp, exp string) (int, error) {
	// parse id from expression
	matches := regex.FindStringSubmatch(exp)
	if len(matches) < 2 {
		return -1, fmt.Errorf("id not found")
	}

	// convert id to integer
	id, _ := strconv.Atoi(matches[1])

	return id, nil
}
