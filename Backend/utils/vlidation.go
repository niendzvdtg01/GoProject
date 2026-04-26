package utils

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

func ValidationPositive(feildName, value string) (int, error) {

	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("Error: %s must be a number", feildName)
	}
	if v <= 0 {
		return 0, fmt.Errorf("%s must be a positive number", feildName)
	}
	return v, nil
}

func ValidationUuid(feildName, value string) (*uuid.UUID, error) {
	uid, err := uuid.Parse(value)
	if err != nil {
		return &uuid.Nil, fmt.Errorf("%s must be uuid", feildName)
	}
	return &uid, nil
}

func ValidationInList(feildName, value string, allowed map[string]bool) error {
	if !allowed[value] {
		return fmt.Errorf("%s must be one of %v", feildName, keys(allowed))
	}
	return nil
}

func keys(m map[string]bool) []string {
	var k []string

	for key := range m {
		k = append(k, key)
	}
	return k
}

func HandleValidatorErrors(err error) gin.H {
	if validationError, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range validationError {
			switch e.Tag() {
			case "gt":
				errors[e.Field()] = e.Field() + "the number must be larger than zero!"
			case "uuid":
				errors[e.Field()] = e.Field() + "the number must be a valid uuid!"
			}
		}
		return gin.H{"error": errors}
	}
	return gin.H{"error": "Yeu cau khong hop le" + err.Error()}
}
