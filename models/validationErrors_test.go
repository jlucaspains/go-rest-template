package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationErrorsSingle(t *testing.T) {
	err := ValidationErrors{
		fmt.Errorf("error 1"),
	}

	assert.Equal(t, "[0]: error 1", err.Error())
}

func TestValidationMultipleErrors(t *testing.T) {
	err := ValidationErrors{
		fmt.Errorf("error 1"),
		fmt.Errorf("error 2"),
	}

	assert.Equal(t, "[0]: error 1\n[1]: error 2", err.Error())
}

func TestValidationErrorsEmpty(t *testing.T) {
	err := ValidationErrors{}

	assert.Equal(t, "", err.Error())
}
