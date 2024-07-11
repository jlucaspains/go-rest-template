package handlers

import (
	"context"
	"encoding/json"
	"goapi-template/auth"
	"goapi-template/db"
	"goapi-template/middlewares"
	"goapi-template/models"
	"goapi-template/util"
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

func New(querier db.Querier) Handlers {
	return Handlers{Queries: querier}
}

func ErrorToHttpResult(err error, ctx context.Context) (int, *models.ErrorResult) {
	slog.Error("Error handled",
		"error", err,
		"traceId", ctx.Value(middlewares.ContextKey("traceId")),
	)

	if vErrs, ok := err.(validator.ValidationErrors); ok {
		out := util.TranslateErrors(vErrs)
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

func GetUser(ctx context.Context) *auth.User {
	if ctx == nil {
		return nil
	}

	if user := ctx.Value(auth.UserKey); user != nil {
		return user.(*auth.User)
	}

	return nil
}

func GetUserEmail(ctx context.Context) string {
	if user := GetUser(ctx); user != nil {
		return user.Email
	}

	return ""
}

func BindJSON(r *http.Request, result any) error {
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
		validateRet := make(models.ValidationErrors, 0)
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

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	result, _ := json.Marshal(data)
	w.Write(result)
}

func Status(w http.ResponseWriter, statusCode int) {
	w.WriteHeader(statusCode)
}

func Param(r *http.Request, key string) string {
	return r.PathValue(key)
}
