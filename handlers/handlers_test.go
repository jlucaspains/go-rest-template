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
	UpdatePersonResult  db.Person
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

func (m *QuerierMock) UpdatePerson(ctx context.Context, arg db.UpdatePersonParams) (db.Person, error) {
	return m.UpdatePersonResult, m.UpdatePersonError
}

func (m *QuerierMock) DeletePerson(ctx context.Context, id int32) (int64, error) {
	return m.DeletePersonResult, m.DeletePersonError
}

func (m *QuerierMock) PingDb(ctx context.Context) (int32, error) {
	return m.PingDbResult, m.PingDbError
}

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

func TestGetUser(t *testing.T) {

	handlers := new(Handlers)
	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))
	body := &auth.User{ID: "test", Name: "test", Email: "email@test.com"}
	newReq := req.WithContext(context.WithValue(req.Context(), auth.UserKey, body))

	user := handlers.GetUser(newReq)

	assert.Equal(t, "test", user.ID)
	assert.Equal(t, "test", user.Name)
	assert.Equal(t, "email@test.com", user.Email)
}

func TestGetUserEmail(t *testing.T) {
	handlers := new(Handlers)
	req, _ := http.NewRequest("GET", "/dummy", bytes.NewReader([]byte("")))
	body := &auth.User{ID: "test", Name: "test", Email: "email@test.com"}
	newReq := req.WithContext(context.WithValue(req.Context(), auth.UserKey, body))

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
	handlers := &Handlers{Queries: querierMock}
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
