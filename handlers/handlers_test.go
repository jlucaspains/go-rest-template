package handlers

import (
	"goapi-template/auth"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
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
	user := handlers.GetUser(new(MockGet))

	assert.Equal(t, "test", user.ID)
	assert.Equal(t, "test", user.Name)
	assert.Equal(t, "email@test.com", user.Email)
}

func TestGetUserEmail(t *testing.T) {
	handlers := new(Handlers)
	email := handlers.GetUserEmail(new(MockGet))

	assert.Equal(t, "email@test.com", email)
}

func TestGetUserEmailEmpty1(t *testing.T) {
	handlers := new(Handlers)
	email := handlers.GetUserEmail(nil)

	assert.Equal(t, "", email)
}

func TestGetUserEmailEmpty2(t *testing.T) {
	handlers := new(Handlers)
	email := handlers.GetUserEmail(new(MockGetNothing))

	assert.Equal(t, "", email)
}
