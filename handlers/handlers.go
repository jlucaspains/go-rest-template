package handlers

import (
	"encoding/json"
	"errors"
	"goapi-template/auth"
	"goapi-template/models"
	"goapi-template/util"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Handlers struct {
	DB *gorm.DB
}

func (h Handlers) ErrorToHttpResult(err error) (int, *models.ErrorResult) {
	if vErrs, ok := err.(validator.ValidationErrors); ok {
		out := util.TranslateErrors(vErrs)
		return http.StatusBadRequest, &models.ErrorResult{Errors: out}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return http.StatusNotFound, &models.ErrorResult{Errors: []string{"Record not found"}}
	} else if errors.Is(err, gorm.ErrDuplicatedKey) {
		return http.StatusConflict, &models.ErrorResult{Errors: []string{"Record duplication detected"}}
	} else if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		// With gorm error translation, most providers will translate
		// unique key errors automatically. The SQLite provider used for
		// testing will not though. This workaround is primarily help unit tests.
		return http.StatusConflict, &models.ErrorResult{Errors: []string{"Record duplication detected"}}
	}

	return http.StatusInternalServerError, &models.ErrorResult{Errors: []string{"Unknown error"}}
}

func (h Handlers) GetUser(r *http.Request) *auth.User {
	if r == nil {
		return nil
	}

	if user := r.Context().Value(auth.UserKey); user != nil {
		return user.(*auth.User)
	}

	return nil
}

func (h Handlers) GetUserEmail(r *http.Request) string {
	if user := h.GetUser(r); user != nil {
		return user.Email
	}

	return ""
}

func (h Handlers) BindJSON(r *http.Request, result any) error {
	err := json.NewDecoder(r.Body).Decode(result)

	if err != nil {
		return err
	}

	validate := validator.New()
	validate.SetTagName("binding")
	value := reflect.ValueOf(result)
	switch value.Kind() {
	case reflect.Ptr:
		return validate.Struct(value.Elem().Interface())
	case reflect.Struct:
		return validate.Struct(result)
	case reflect.Slice, reflect.Array:
		count := value.Len()
		validateRet := make(models.SliceValidationError, 0)
		for i := 0; i < count; i++ {
			if err := validate.Struct(value.Index(i).Interface()); err != nil {
				validateRet = append(validateRet, err)
			}
		}
		if len(validateRet) == 0 {
			return nil
		}
		return validateRet
	default:
		return nil
	}
}

func (h Handlers) JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	result, _ := json.Marshal(data)
	w.Write(result)
}

func (h Handlers) Status(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

func (h Handlers) Param(r *http.Request, key string) string {
	vars := mux.Vars(r)

	return vars[key]
}
