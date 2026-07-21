package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/adservice/internal/app/dto"
	"github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/adservice/pkg/errs"
)

type RejectAdUC struct {
	ad port.AdRepository
}

func NewRejectAdUC(ad port.AdRepository) *RejectAdUC {
	return &RejectAdUC{ad: ad}
}

func (uc *RejectAdUC) Execute(ctx context.Context, in dto.RejectAdInput) (dto.RejectAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.RejectAdOutput{Success: false}, errs.ErrInvalidAdID
		}
		return dto.RejectAdOutput{Success: false}, errs.Wrap(
			errs.ErrGetAdDB, err,
		)
	}

	// Check if current user can reject this ad
	if ad.SellerID() != in.SellerID {
		return dto.RejectAdOutput{Success: false}, errs.ErrAccessDenied
	}

	// Reject
	err = ad.Reject()
	if err != nil {
		return dto.RejectAdOutput{Success: false}, errs.ErrCannotReject
	}

	// Update in db
	err = uc.ad.UpdateStatus(ctx, ad)
	if err != nil {
		return dto.RejectAdOutput{Success: false}, errs.Wrap(
			errs.ErrUpdateAdStatusDB, err,
		)
	}

	// Response
	return dto.RejectAdOutput{Success: true}, nil
}
