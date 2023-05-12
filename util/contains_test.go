package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	collection := []string{
		"test1", "test2", "test3",
	}

	result := Contains(collection, "test1")

	assert.True(t, result)
}
