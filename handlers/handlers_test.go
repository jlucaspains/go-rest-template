package handlers

import (
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
