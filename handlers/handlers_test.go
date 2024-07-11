package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"goapi-template/auth"
	"goapi-template/db"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

type QuerierMock struct {
	GetPeopleResult     []db.Person
	GetPeopleError      error
	GetPersonByIdResult db.Person
	GetPersonByIdError  error
	InsertPersonResult  db.Person
	InsertPersonError   error
	UpdatePersonResult  int64
	UpdatePersonError   error
	DeletePersonResult  int64
	DeletePersonError   error
	PingDbResult        int32
	PingDbError         error
}

func (m *QuerierMock) GetPeople(ctx context.Context) ([]db.Person, error) {
	return m.GetPeopleResult, m.GetPeopleError
}

func (m *QuerierMock) GetPersonById(ctx context.Context, id int32) (db.Person, error) {
	return m.GetPersonByIdResult, m.GetPersonByIdError
}

func (m *QuerierMock) InsertPerson(ctx context.Context, arg db.InsertPersonParams) (db.Person, error) {
	return m.InsertPersonResult, m.InsertPersonError
}

func (m *QuerierMock) UpdatePerson(ctx context.Context, arg db.UpdatePersonParams) (int64, error) {
	return m.UpdatePersonResult, m.UpdatePersonError
}

func (m *QuerierMock) DeletePerson(ctx context.Context, id int32) (int64, error) {
	return m.DeletePersonResult, m.DeletePersonError
}

func (m *QuerierMock) PingDb(ctx context.Context) (int32, error) {
	return m.PingDbResult, m.PingDbError
}

func TestGetUser(t *testing.T) {

	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))
	body := &auth.User{ID: "test", Name: "test", Email: "email@test.com"}
	newReq := req.WithContext(context.WithValue(req.Context(), auth.UserKey, body))

	user := GetUser(newReq.Context())

	assert.Equal(t, "test", user.ID)
	assert.Equal(t, "test", user.Name)
	assert.Equal(t, "email@test.com", user.Email)
}

func TestGetUserEmail(t *testing.T) {
	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))
	body := &auth.User{ID: "test", Name: "test", Email: "email@test.com"}
	newReq := req.WithContext(context.WithValue(req.Context(), auth.UserKey, body))

	email := GetUserEmail(newReq.Context())

	assert.Equal(t, "email@test.com", email)
}

func TestGetUserEmailEmpty1(t *testing.T) {
	email := GetUserEmail(context.TODO())

	assert.Equal(t, "", email)
}

func TestGetUserEmailEmpty2(t *testing.T) {
	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))

	email := GetUserEmail(req.Context())

	assert.Equal(t, "", email)
}

func makeRequest[K any | []any](router *http.ServeMux, method string, url string, body any) (code int, respBody *K, headers http.Header, err error) {
	inputBody := ""

	if body != nil {
		inputBodyJson, _ := json.Marshal(body)
		inputBody = string(inputBodyJson)
	}

	req, _ := http.NewRequest(method, url, bytes.NewReader([]byte(inputBody)))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	result := new(K)

	switch any(result).(type) {
	case *string:
		// do nothing as we don't care about string
		result = nil
	default:
		err = json.Unmarshal(rr.Body.Bytes(), &result)
	}

	return rr.Code, result, rr.Result().Header, err
}

func setup(querierMock *QuerierMock) *http.ServeMux {
	router := http.NewServeMux()
	handlers := New(querierMock)
	router.Handle("GET /person/{id}", mockAuthMiddleware(http.HandlerFunc(handlers.GetPerson)))
	router.Handle("PUT /person/{id}", mockAuthMiddleware(http.HandlerFunc(handlers.PutPerson)))
	router.Handle("POST /person", mockAuthMiddleware(http.HandlerFunc(handlers.PostPerson)))
	router.Handle("DELETE /person/{id}", mockAuthMiddleware(http.HandlerFunc(handlers.DeletePerson)))
	router.HandleFunc("GET /health", handlers.GetHealth)

	godotenv.Load("../.testing.env")

	return router
}

func mockAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := &auth.User{ID: "Test", Name: "Test", Email: "mail@test.com"}
		r = r.WithContext(context.WithValue(ctx, auth.UserKey, user))

		next.ServeHTTP(w, r)
	})
}

func TestErrorTranslationSuccess(t *testing.T) {
	type TestStruct struct {
		Req   string `validate:"required"`
		Lt    string `validate:"required,lt=10"`
		Lte   string `validate:"required,lte=1"`
		Gt    int    `validate:"required,gt=1"`
		Gte   int    `validate:"required,gte=10"`
		Min   string `validate:"min=10"`
		Max   string `validate:"max=9"`
		Alpha string `validate:"alpha"`
	}

	user := TestStruct{
		Req:   "",
		Lt:    "0123456789",
		Lte:   "012345678",
		Gt:    1,
		Gte:   1,
		Min:   "012345678",
		Max:   "0123456789",
		Alpha: "0123456789",
	}

	validate := validator.New()
	err := validate.Struct(user)
	code, result := ErrorToHttpResult(err, context.Background())

	assert.Equal(t, http.StatusBadRequest, code)

	assert.Len(t, result.Errors, 8)
	assert.Equal(t, "Req is required", result.Errors[0])
	assert.Equal(t, "Lt should be less than 10", result.Errors[1])
	assert.Equal(t, "Lte should be less than or equal to 1", result.Errors[2])
	assert.Equal(t, "Gt should be greater than 1", result.Errors[3])
	assert.Equal(t, "Gte should be greater than or equal to 10", result.Errors[4])
	assert.Equal(t, "Min should have minimum length of 10", result.Errors[5])
	assert.Equal(t, "Max should have maximum length of 9", result.Errors[6])
	assert.Equal(t, "Alpha should contain alpha characters only", result.Errors[7])
}

func TestErrorTranslationServerError(t *testing.T) {
	code, result := ErrorToHttpResult(fmt.Errorf("Something went wrong"), context.Background())
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.Equal(t, "Unknown error", result.Errors[0])
}
