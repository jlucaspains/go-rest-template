package handlers

import (
	"errors"
	"goapi-template/models"
	"goapi-template/util"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Handlers struct {
	DB         *gorm.DB
}

func (h Handlers) ErrorToHttpResult(err error) (int, *models.ErrorResult) {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		out := util.TranslateErrors(vErrs)
		return http.StatusBadRequest, &models.ErrorResult{Errors: out}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, &models.ErrorResult{Errors: []string{"Record not found"}}
	} else if errors.Is(err, gorm.ErrDuplicatedKey) {
		return http.StatusNotFound, &models.ErrorResult{Errors: []string{"Record duplication detected"}}
	}

	return http.StatusInternalServerError, &models.ErrorResult{Errors: []string{"Invalid request body"}}
}
