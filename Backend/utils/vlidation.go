package utils

import "fmt"

func ValidationRequire(feildName, value string) error {
	if value == "" {
		return fmt.Errorf("%s is require", feildName)
	}
	return nil
}
