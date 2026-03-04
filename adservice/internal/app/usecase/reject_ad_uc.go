package usecase

import (
	"ads/adservice/internal/app/dto"
	"ads/adservice/internal/app/uc_errors"
	"ads/adservice/internal/domain/port"
	"ads/pkg/errs"
	"context"
	"errors"
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
		if errors.Is(err, errs.ErrObjectNotFound) {
			return dto.RejectAdOutput{Success: false}, uc_errors.ErrInvalidAdID
		}
		return dto.RejectAdOutput{Success: false}, uc_errors.Wrap(
			uc_errors.ErrGetAdDB, err,
		)
	}

	// Check if current user can reject this ad
	if ad.SellerID() != in.SellerID {
		return dto.RejectAdOutput{Success: false}, uc_errors.ErrAccessDenied
	}

	// Reject
	err = ad.Reject()
	if err != nil {
		return dto.RejectAdOutput{Success: false}, uc_errors.ErrCannotReject
	}

	// Update in db
	err = uc.ad.UpdateStatus(ctx, ad)
	if err != nil {
		return dto.RejectAdOutput{Success: false}, uc_errors.Wrap(
			uc_errors.ErrUpdateAdStatusDB, err,
		)
	}

	// Response
	return dto.RejectAdOutput{Success: true}, nil
}
