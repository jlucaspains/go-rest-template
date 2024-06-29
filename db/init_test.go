package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBInitBadConnectionString(t *testing.T) {
	err := Init("Justbad")

	assert.Error(t, err)
}
