package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/app/usecase"
	"github.com/maket12/ads-service/adservice/internal/domain/port/mocks"
	"github.com/maket12/ads-service/userservice/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateAdUC_Execute(t *testing.T) {
	type adapter struct {
		ad    *mocks.MockAdRepository
		media *mocks.MockMediaRepository
	}

	type testCase struct {
		name          string
		input         dto.CreateAdInput
		mockBehaviour func(a adapter)
		expectErr     error
	}

	sellerID := uuid.New()
	title := gofakeit.ProductName()
	description := utils.VPtr(gofakeit.ProductDescription())
	price := int64(gofakeit.Price(1, 1000))
	images := []string{gofakeit.URL(), gofakeit.URL()}

	var tests = []testCase{
		{
			name: "Success",
			input: dto.CreateAdInput{
				SellerID:    sellerID,
				Title:       title,
				Description: description,
				Price:       price,
				Images:      images,
			},
			mockBehaviour: func(a adapter) {
				a.ad.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Ad")).
					Return(nil)

				a.media.EXPECT().
					Save(mock.Anything, mock.AnythingOfType("uuid.UUID"), images).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - invalid input",
			input: dto.CreateAdInput{
				SellerID:    sellerID,
				Title:       "",
				Description: description,
				Price:       price,
				Images:      images,
			},
			mockBehaviour: func(a adapter) {},
			expectErr:     ucerrs.ErrInvalidInput,
		},
		{
			name: "Failure - create ad db error",
			input: dto.CreateAdInput{
				SellerID:    sellerID,
				Title:       title,
				Description: description,
				Price:       price,
				Images:      images,
			},
			mockBehaviour: func(a adapter) {
				a.ad.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Ad")).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrCreateAdDB,
		},
		{
			name: "Failure - save images db error",
			input: dto.CreateAdInput{
				SellerID:    sellerID,
				Title:       title,
				Description: description,
				Price:       price,
				Images:      images,
			},
			mockBehaviour: func(a adapter) {
				a.ad.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*model.Ad")).
					Return(nil)

				a.media.EXPECT().
					Save(mock.Anything, mock.AnythingOfType("uuid.UUID"), images).
					Return(errors.New("db error"))
			},
			expectErr: ucerrs.ErrSaveImagesDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adRepo := mocks.NewMockAdRepository(t)
			mediaRepo := mocks.NewMockMediaRepository(t)

			tt.mockBehaviour(adapter{
				ad:    adRepo,
				media: mediaRepo,
			})

			uc := usecase.NewCreateAdUC(mocks.FakeTxManager{}, adRepo, mediaRepo)

			out, err := uc.Execute(context.Background(), tt.input)

			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, out.AdID)
			}
		})
	}
}
