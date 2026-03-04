package usecase

import (
	"ads/adservice/internal/app/dto"
	"ads/adservice/internal/app/uc_errors"
	"ads/adservice/internal/domain/port"
	"ads/pkg/errs"
	"context"
	"errors"
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
		if errors.Is(err, errs.ErrObjectNotFound) {
			return dto.PublishAdOutput{Success: false}, uc_errors.ErrInvalidAdID
		}
		return dto.PublishAdOutput{Success: false}, uc_errors.Wrap(
			uc_errors.ErrGetAdDB, err,
		)
	}

	// Check if current user can publish this ad
	if ad.SellerID() != in.SellerID {
		return dto.PublishAdOutput{Success: false}, uc_errors.ErrAccessDenied
	}

	// Publish
	err = ad.Publish()
	if err != nil {
		return dto.PublishAdOutput{Success: false}, uc_errors.ErrCannotPublish
	}

	// Update in db
	err = uc.ad.UpdateStatus(ctx, ad)
	if err != nil {
		return dto.PublishAdOutput{Success: false}, uc_errors.Wrap(
			uc_errors.ErrUpdateAdStatusDB, err,
		)
	}

	// Response
	return dto.PublishAdOutput{Success: true}, nil
}
