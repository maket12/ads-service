package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/searchservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/usecase"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/port/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAdIndexUC_Execute(t *testing.T) {
	type testCase struct {
		name          string
		input         dto.CreateAdIndexInput
		mockBehaviour func(a *mocks.MockAdIndexRepository)
		expectErr     error
	}

	id := gofakeit.UUID()
	title := gofakeit.ProductName()
	description := gofakeit.ProductDescription()
	price := int64(gofakeit.Price(1, 1000))
	category := model.CategoryHome.String()
	mainImage := gofakeit.URL()

	var tests = []testCase{
		{
			name: "Success",
			input: dto.CreateAdIndexInput{
				ID:          id,
				Title:       title,
				Description: description,
				Price:       price,
				Category:    category,
				MainImage:   mainImage,
			},
			mockBehaviour: func(a *mocks.MockAdIndexRepository) {
				a.EXPECT().
					Index(mock.Anything, mock.AnythingOfType("*model.AdIndex")).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - invalid input",
			input: dto.CreateAdIndexInput{
				ID:          id,
				Title:       "",
				Description: description,
				Price:       price,
				Category:    category,
				MainImage:   mainImage,
			},
			mockBehaviour: func(_ *mocks.MockAdIndexRepository) {},
			expectErr:     ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - index es error",
			input: dto.CreateAdIndexInput{
				ID:          id,
				Title:       title,
				Description: description,
				Price:       price,
				Category:    category,
				MainImage:   mainImage,
			},
			mockBehaviour: func(a *mocks.MockAdIndexRepository) {
				a.EXPECT().
					Index(mock.Anything, mock.AnythingOfType("*model.AdIndex")).
					Return(errors.New("es error"))
			},
			expectErr: ucerrs.ErrIndexAdES,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adIndexRepo := mocks.NewMockAdIndexRepository(t)

			tt.mockBehaviour(adIndexRepo)

			uc := usecase.NewCreateAdIndexUC(adIndexRepo)

			err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
