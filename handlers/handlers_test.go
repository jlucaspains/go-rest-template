package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"goapi-template/auth"
	"goapi-template/db"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestErrorTranslationSuccess(t *testing.T) {
	type User struct {
		Username string `validate:"required"`
		Tagline  string `validate:"required,lt=10"`
		Tagline2 int    `validate:"required,gt=1"`
	}

	user := User{
		Username: "Joeybloggs",
		Tagline:  "Works",
		Tagline2: 1,
	}

	validate := validator.New()
	err := validate.Struct(user)
	handlers := new(Handlers)
	code, result := handlers.ErrorToHttpResult(err)

	assert.Equal(t, http.StatusBadRequest, code)

	assert.Len(t, result.Errors, 1)
	assert.Equal(t, "Tagline2 should be greater than 1", result.Errors[0])
}

type MockGet struct{}

func (c MockGet) Get(key string) (value any, exists bool) {
	if key == "User" {
		return &auth.User{ID: "test", Name: "test", Email: "email@test.com"}, true
	}

	return nil, false
}

type MockGetNothing struct{}

func (c MockGetNothing) Get(key string) (value any, exists bool) {
	return nil, false
}

func TestGetUser(t *testing.T) {

	handlers := new(Handlers)
	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))
	body := &auth.User{ID: "test", Name: "test", Email: "email@test.com"}
	newReq := req.WithContext(context.WithValue(req.Context(), "User", body))

	user := handlers.GetUser(newReq)

	assert.Equal(t, "test", user.ID)
	assert.Equal(t, "test", user.Name)
	assert.Equal(t, "email@test.com", user.Email)
}

func TestGetUserEmail(t *testing.T) {
	handlers := new(Handlers)
	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))
	body := &auth.User{ID: "test", Name: "test", Email: "email@test.com"}
	newReq := req.WithContext(context.WithValue(req.Context(), "User", body))

	email := handlers.GetUserEmail(newReq)

	assert.Equal(t, "email@test.com", email)
}

func TestGetUserEmailEmpty1(t *testing.T) {
	handlers := new(Handlers)
	email := handlers.GetUserEmail(nil)

	assert.Equal(t, "", email)
}

func TestGetUserEmailEmpty2(t *testing.T) {
	handlers := new(Handlers)
	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))

	email := handlers.GetUserEmail(req)

	assert.Equal(t, "", email)
}

func makeRequest[K any | []any](router *mux.Router, method string, url string, body any) (code int, respBody *K, err error) {
	inputBody := ""

	if body != nil {
		inputBodyJson, _ := json.Marshal(body)
		inputBody = string(inputBodyJson)
	}

	req, _ := http.NewRequest(method, url, bytes.NewReader([]byte(inputBody)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	result := new(K)
	err = json.Unmarshal(rr.Body.Bytes(), &result)

	return rr.Code, result, err
}

func setup(migrate bool, useAuthMiddleware bool) (*mux.Router, *gorm.DB) {
	router := mux.NewRouter()

	godotenv.Load("../.testing.env")
	db, err := db.Init("sqlite", ":memory:", migrate)

	if err != nil {
		panic(err)
	}

	if useAuthMiddleware {
		authMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				user := &auth.User{ID: "Test", Name: "Test", Email: "mail@test.com"}
				req := r.WithContext(context.WithValue(ctx, "User", user))
				next.ServeHTTP(w, req)
			})
		}
		router.Use(authMiddleware)
	}

	return router, db
}
