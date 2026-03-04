package usecase

import (
	"context"
	"errors"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/adservice/internal/app/errs"
	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
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
			return dto.RejectAdOutput{Success: false}, ucerrs.ErrInvalidAdID
		}
		return dto.RejectAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can reject this ad
	if ad.SellerID() != in.SellerID {
		return dto.RejectAdOutput{Success: false}, ucerrs.ErrAccessDenied
	}

	// Reject
	err = ad.Reject()
	if err != nil {
		return dto.RejectAdOutput{Success: false}, ucerrs.ErrCannotReject
	}

	// Update in db
	err = uc.ad.UpdateStatus(ctx, ad)
	if err != nil {
		return dto.RejectAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrUpdateAdStatusDB, err,
		)
	}

	// Response
	return dto.RejectAdOutput{Success: true}, nil
}
