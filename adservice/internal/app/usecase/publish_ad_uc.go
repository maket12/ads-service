package usecase

import (
	"context"
	"errors"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
)

type PublishAdUC struct {
	ad port.AdRepository
}

func NewPublishAdUC(ad port.AdRepository) *PublishAdUC {
	return &PublishAdUC{ad: ad}
}

func (uc *PublishAdUC) Execute(ctx context.Context, in dto.PublishAdInput) (dto.PublishAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.PublishAdOutput{Success: false}, ucerrs.ErrInvalidAdID
		}
		return dto.PublishAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can publish this ad
	if ad.SellerID() != in.SellerID {
		return dto.PublishAdOutput{Success: false}, ucerrs.ErrAccessDenied
	}

	// Publish
	err = ad.Publish()
	if err != nil {
		return dto.PublishAdOutput{Success: false}, ucerrs.ErrCannotPublish
	}

	// Update in db
	err = uc.ad.UpdateStatus(ctx, ad)
	if err != nil {
		return dto.PublishAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrUpdateAdStatusDB, err,
		)
	}

	// Response
	return dto.PublishAdOutput{Success: true}, nil
}
