package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

func ValidationRequire(feildName, value string) error {
	if value == "" {
		return fmt.Errorf("%s is require", feildName)
	}
	return nil
}

func ValidationStringLength(feildName, value string, min, max int) error {
	l := len(value)
	if l < min || l > max {
		return fmt.Errorf("%s  must be between %d and %d chararcters", feildName, min, max)
	}
	return nil
}

func ValidationCharacter(erorMessage, value string, re *regexp.Regexp) error {
	if !re.MatchString(value) {
		return fmt.Errorf("%s", erorMessage)
	}
	return nil
}

func ValidationPositive(idStr string) error {

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("Error:")
	}
	if id <= 0 {
		return fmt.Errorf("")
	}
	return nil
}
