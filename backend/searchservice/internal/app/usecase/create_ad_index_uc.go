package usecase

import (
	"context"

	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/searchservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/port"
)

type CreateAdIndexUC struct {
	adIndex port.AdIndexRepository
}

func NewCreateAdIndexUC(adIndex port.AdIndexRepository) *CreateAdIndexUC {
	return &CreateAdIndexUC{adIndex: adIndex}
}

func (uc *CreateAdIndexUC) Execute(ctx context.Context, in dto.CreateAdIndexInput) error {
	adIndex, err := model.NewAdIndex(
		in.ID, in.Title,
		in.Description, in.Price,
		in.Category, in.MainImage,
	)
	if err != nil {
		return ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
	}

	if err = uc.adIndex.Index(ctx, adIndex); err != nil {
		return ucerrs.Wrap(ucerrs.ErrIndexAdES, err)
	}

	return nil
}
