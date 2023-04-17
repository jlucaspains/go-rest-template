package util

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func TranslateErrors(err validator.ValidationErrors) []string {
	out := make([]string, len(err))
	for i, fe := range err {
		out[i] = getValidationErrorMsg(fe)
	}
	return out
}

func getValidationErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "lte":
		return fmt.Sprintf("%s should be less than or equal to %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("%s should be less than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("%s should be greater than or equal to %s", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("%s should be greater than %s", fe.Field(), fe.Param())
	case "min":
		return fmt.Sprintf("%s should have minimum length of %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s should have maximum length of %s", fe.Field(), fe.Param())
	case "alpha":
		return fmt.Sprintf("%s should contain alpha characters only", fe.Field())
	}
	return "Unknown error"
}
