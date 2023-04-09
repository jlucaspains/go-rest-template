package util

import (
	"errors"
	"fmt"
	"goapi-template/models"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

func ErrorToHttpResult(err error) (int, *models.ErrorResult) {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		out := TranslateErrors(vErrs)
		return http.StatusBadRequest, &models.ErrorResult{Errors: out}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, &models.ErrorResult{Errors: []string{"Record not found"}}
	} else if errors.Is(err, gorm.ErrDuplicatedKey) {
		return http.StatusNotFound, &models.ErrorResult{Errors: []string{"Record duplication detected"}}
	}

	return http.StatusInternalServerError, &models.ErrorResult{Errors: []string{"Invalid request body"}}
}

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
		return fmt.Sprintf("%s should be less than %s", fe.Field(), fe.Param())
	case "gte":
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
