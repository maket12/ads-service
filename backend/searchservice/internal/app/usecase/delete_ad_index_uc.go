package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/backend/searchservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/searchservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/searchservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/backend/searchservice/pkg/errs"
)

type DeleteAdIndexUC struct {
	adIndex port.AdIndexRepository
}

func NewDeleteAdIndexUC(adIndex port.AdIndexRepository) *DeleteAdIndexUC {
	return &DeleteAdIndexUC{adIndex: adIndex}
}

func (uc *DeleteAdIndexUC) Execute(ctx context.Context, in dto.DeleteAdIndexInput) error {
	if in.ID == "" {
		return ucerrs.ErrInvalidAdIndex
	}

	if err := uc.adIndex.Delete(ctx, in.ID); err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return ucerrs.ErrAdIndexNotFound
		}
		return ucerrs.Wrap(ucerrs.ErrDeleteAdIndexES, err)
	}

	return nil
}
