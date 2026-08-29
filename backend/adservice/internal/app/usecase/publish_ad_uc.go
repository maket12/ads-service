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

type PublishAdUC struct {
	trManager trm.Manager
	ad        port.AdRepository
	publisher port.AdPublisher
}

func NewPublishAdUC(
	trManager trm.Manager,
	ad port.AdRepository,
	publisher port.AdPublisher,
) *PublishAdUC {
	return &PublishAdUC{
		trManager: trManager,
		ad:        ad,
		publisher: publisher,
	}
}

func (uc *PublishAdUC) Execute(ctx context.Context, in dto.PublishAdInput) (dto.PublishAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.PublishAdOutput{}, ucerrs.ErrAdNotFound
		}
		return dto.PublishAdOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Publish
	err = ad.Publish()
	if err != nil {
		return dto.PublishAdOutput{Success: false}, ucerrs.ErrCannotPublish
	}

	// Update in db and publish in queue
	if err = uc.trManager.Do(ctx, func(txCtx context.Context) error {
		if updErr := uc.ad.Update(txCtx, ad); updErr != nil {
			return ucerrs.Wrap(ucerrs.ErrUpdateAdDB, err)
		}

		if publishErr := uc.publisher.PublishAdPublished(txCtx, ad); publishErr != nil {
			return ucerrs.Wrap(ucerrs.ErrPublishAdPublishedEvent, publishErr)
		}

		return nil
	}); err != nil {
		return dto.PublishAdOutput{}, err
	}

	// Response
	return dto.PublishAdOutput{Success: true}, nil
}
