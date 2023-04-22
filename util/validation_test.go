package util

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestTranslateErrorSuccess(t *testing.T) {
	type User struct {
		Username         string `validate:"required"`
		LessThanEqual    string `validate:"lte=10"`
		LessThan         string `validate:"lt=10"`
		GreaterThanEqual int    `validate:"gte=10"`
		GreaterThan      int    `validate:"gt=10"`
		Min              int    `validate:"min=10"`
		Max              int    `validate:"max=10"`
		Alpha            string `validate:"alpha"`
	}

	user := User{
		Username:         "",
		LessThan:         "VeryLongSoItFails",
		LessThanEqual:    "VeryLongSoItFails",
		GreaterThan:      1,
		GreaterThanEqual: 1,
		Min:              9,
		Max:              11,
		Alpha:            "123",
	}

	validate := validator.New()
	err := validate.Struct(user)
	errors, _ := err.(validator.ValidationErrors)

	result := TranslateErrors(errors)

	assert.Len(t, result, 8)
	assert.Equal(t, "Username is required", result[0])
	assert.Equal(t, "LessThanEqual should be less than or equal to 10", result[1])
	assert.Equal(t, "LessThan should be less than 10", result[2])
	assert.Equal(t, "GreaterThanEqual should be greater than or equal to 10", result[3])
	assert.Equal(t, "GreaterThan should be greater than 10", result[4])
	assert.Equal(t, "Min should have minimum length of 10", result[5])
	assert.Equal(t, "Max should have maximum length of 10", result[6])
	assert.Equal(t, "Alpha should contain alpha characters only", result[7])
}
