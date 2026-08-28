package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	pkgerrs "github.com/maket12/ads-service/backend/adservice/pkg/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/searchservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/usecase"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/port/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteAdIndexUC_Execute(t *testing.T) {
	type testCase struct {
		name          string
		input         dto.DeleteAdIndexInput
		mockBehaviour func(a *mocks.MockAdIndexRepository)
		expectErr     error
	}

	id := gofakeit.UUID()

	var tests = []testCase{
		{
			name: "Success",
			input: dto.DeleteAdIndexInput{
				ID: id,
			},
			mockBehaviour: func(a *mocks.MockAdIndexRepository) {
				a.EXPECT().
					Delete(mock.Anything, id).
					Return(nil)
			},
			expectErr: nil,
		},
		{
			name: "Failure - empty id",
			input: dto.DeleteAdIndexInput{
				ID: "",
			},
			mockBehaviour: func(_ *mocks.MockAdIndexRepository) {},
			expectErr:     ucerrs.ErrInvalidAdIndex,
		},
		{
			name: "Failure - not found",
			input: dto.DeleteAdIndexInput{
				ID: id,
			},
			mockBehaviour: func(a *mocks.MockAdIndexRepository) {
				a.EXPECT().
					Delete(mock.Anything, id).
					Return(pkgerrs.ErrObjectNotFound)
			},
			expectErr: ucerrs.ErrAdIndexNotFound,
		},
		{
			name: "Failure - delete es error",
			input: dto.DeleteAdIndexInput{
				ID: id,
			},
			mockBehaviour: func(a *mocks.MockAdIndexRepository) {
				a.EXPECT().
					Delete(mock.Anything, id).
					Return(errors.New("es error"))
			},
			expectErr: ucerrs.ErrDeleteAdIndexES,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adIndexRepo := mocks.NewMockAdIndexRepository(t)

			tt.mockBehaviour(adIndexRepo)

			uc := usecase.NewDeleteAdIndexUC(adIndexRepo)

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
