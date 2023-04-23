package handlers

import (
	"errors"
	"goapi-template/auth"
	"goapi-template/models"
	"goapi-template/util"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Handlers struct {
	DB *gorm.DB
}

type HaveGet interface {
	Get(key string) (value any, exists bool)
}

func (h Handlers) ErrorToHttpResult(err error) (int, *models.ErrorResult) {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		out := util.TranslateErrors(vErrs)
		return http.StatusBadRequest, &models.ErrorResult{Errors: out}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, &models.ErrorResult{Errors: []string{"Record not found"}}
	} else if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return http.StatusConflict, &models.ErrorResult{Errors: []string{"Record duplication detected"}}
	}

	return http.StatusInternalServerError, &models.ErrorResult{Errors: []string{"Invalid request body"}}
}

func (h Handlers) GetUser(c HaveGet) *auth.User {
	if c == nil {
		return nil
	}

	if user, _ := c.Get("User"); user != nil {
		return user.(*auth.User)
	}

	return nil
}

func (h Handlers) GetUserEmail(c HaveGet) string {
	if user := h.GetUser(c); user != nil {
		return user.Email
	}

	return ""
}
