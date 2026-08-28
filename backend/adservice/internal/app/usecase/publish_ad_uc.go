package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
)

type PublishAdUC struct{ ad port.AdRepository }

func NewPublishAdUC(ad port.AdRepository) *PublishAdUC {
	return &PublishAdUC{ad: ad}
}

func (uc *PublishAdUC) Execute(ctx context.Context, in dto.PublishAdInput) (dto.PublishAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.PublishAdOutput{Success: false}, ucerrs.ErrAdNotFound
		}
		return dto.PublishAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Publish
	err = ad.Publish()
	if err != nil {
		return dto.PublishAdOutput{Success: false}, ucerrs.ErrCannotPublish
	}

	// Update in db
	err = uc.ad.Update(ctx, ad)
	if err != nil {
		return dto.PublishAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrUpdateAdDB, err,
		)
	}

	// Response
	return dto.PublishAdOutput{Success: true}, nil
}
