package handlers

import (
	"encoding/json"
	"fmt"
	"goapi-template/auth"
	"goapi-template/db"
	"goapi-template/models"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handlers struct {
	Queries db.Querier
}

func (h Handlers) ErrorToHttpResult(err error) (int, *models.ErrorResult) {
	slog.Error("Error handled", "error", err)

	if vErrs, ok := err.(validator.ValidationErrors); ok {
		out := translateErrors(vErrs)
		return http.StatusBadRequest, &models.ErrorResult{Errors: out}
	}

	if err == pgx.ErrNoRows {
		return http.StatusNotFound, nil
	}

	if dbError, ok := err.(*pgconn.PgError); ok {
		if dbError.Code == "23505" {
			return http.StatusConflict, &models.ErrorResult{Errors: []string{"Record duplication detected"}}
		}
	}

	return http.StatusInternalServerError, &models.ErrorResult{Errors: []string{"Unknown error"}}
}

func translateErrors(err validator.ValidationErrors) []string {
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
	return r.PathValue(key)
}
