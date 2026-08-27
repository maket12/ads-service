package model_test

import (
	"testing"

	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func TestNewCategory(t *testing.T) {
	validVal := model.CategoryHome.String()
	invalidVal := "unknown"

	category, err := model.NewCategory(validVal)
	assert.NoError(t, err)
	assert.Equal(t, validVal, category.String())

	category, err = model.NewCategory(invalidVal)
	assert.Error(t, err)
}
