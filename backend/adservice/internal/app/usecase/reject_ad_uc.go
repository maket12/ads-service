package usecase

import (
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
)

type RejectAdUC struct {
	trManager trm.Manager
	ad        port.AdRepository
	publisher port.AdPublisher
}

func NewRejectAdUC(
	trManager trm.Manager,
	ad port.AdRepository,
	publisher port.AdPublisher,
) *RejectAdUC {
	return &RejectAdUC{
		trManager: trManager,
		ad:        ad,
		publisher: publisher,
	}
}

func (uc *RejectAdUC) Execute(ctx context.Context, in dto.RejectAdInput) (dto.RejectAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.RejectAdOutput{}, ucerrs.ErrAdNotFound
		}
		return dto.RejectAdOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Reject
	err = ad.Reject()
	if err != nil {
		return dto.RejectAdOutput{}, ucerrs.ErrCannotReject
	}

	// Update in db and publish the event in the queue
	if err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
		if updErr := uc.ad.Update(txCtx, ad); updErr != nil {
			return ucerrs.Wrap(ucerrs.ErrUpdateAdDB, err)
		}

		if publishErr := uc.publisher.PublishAdRejected(txCtx, ad.ID()); publishErr != nil {
			return ucerrs.Wrap(ucerrs.ErrPublishAdRejected, publishErr)
		}

		return nil
	}); err != nil {
		return dto.RejectAdOutput{}, err
	}

	// Response
	return dto.RejectAdOutput{Success: true}, nil
}
