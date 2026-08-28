package usecase

import (
	"context"
	"errors"

	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
)

type RejectAdUC struct{ ad port.AdRepository }

func NewRejectAdUC(ad port.AdRepository) *RejectAdUC {
	return &RejectAdUC{ad: ad}
}

func (uc *RejectAdUC) Execute(ctx context.Context, in dto.RejectAdInput) (dto.RejectAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.RejectAdOutput{Success: false}, ucerrs.ErrAdNotFound
		}
		return dto.RejectAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Reject
	err = ad.Reject()
	if err != nil {
		return dto.RejectAdOutput{Success: false}, ucerrs.ErrCannotReject
	}

	// Update in db
	err = uc.ad.Update(ctx, ad)
	if err != nil {
		return dto.RejectAdOutput{Success: false}, ucerrs.Wrap(
			ucerrs.ErrUpdateAdDB, err,
		)
	}

	// Response
	return dto.RejectAdOutput{Success: true}, nil
}
