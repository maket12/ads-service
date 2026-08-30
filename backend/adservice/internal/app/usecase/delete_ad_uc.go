package usecase

import (
	"context"
	"errors"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	ucerrs "github.com/maket12/ads-service/backend/adservice/internal/app/errs"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/port"
	pkgerrs "github.com/maket12/ads-service/backend/authservice/pkg/errs"
)

type DeleteAdUC struct {
	trManager trm.Manager
	ad        port.AdRepository
	media     port.MediaRepository
	publisher port.AdPublisher
}

func NewDeleteAdUC(
	trManager trm.Manager,
	ad port.AdRepository,
	media port.MediaRepository,
	publisher port.AdPublisher,
) *DeleteAdUC {
	return &DeleteAdUC{
		trManager: trManager,
		ad:        ad,
		media:     media,
		publisher: publisher,
	}
}

func (uc *DeleteAdUC) Execute(ctx context.Context, in dto.DeleteAdInput) (dto.DeleteAdOutput, error) {
	// Get from db
	ad, err := uc.ad.Get(ctx, in.AdID)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return dto.DeleteAdOutput{}, ucerrs.ErrAdNotFound
		}
		return dto.DeleteAdOutput{}, ucerrs.Wrap(
			ucerrs.ErrGetAdDB, err,
		)
	}

	// Check if current user can delete this ad
	if ad.SellerID() != in.SellerID {
		return dto.DeleteAdOutput{}, ucerrs.ErrAccessDenied
	}

	// Check if the ad has been already deleted
	if ad.IsDeleted() {
		return dto.DeleteAdOutput{}, ucerrs.ErrCannotDelete
	}

	// Scenario №1: Delete ad data in database (if not published yet)
	if ad.IsOnModeration() && ad.UpdatedAt() == nil {
		if err = uc.hardDelete(ctx, ad); err != nil {
			return dto.DeleteAdOutput{}, err
		}
	} else {
		// Scenario №2: Update status in database (deleted) and publish the event in the queue
		if err = uc.softDeleteAndPublish(ctx, ad); err != nil {
			return dto.DeleteAdOutput{}, err
		}
	}

	// Response
	return dto.DeleteAdOutput{Success: true}, nil
}

// hardDelete removes the ad and its media completely from the database.
// Used when the ad hasn't been published yet, so there's nothing to announce.
func (uc *DeleteAdUC) hardDelete(ctx context.Context, ad *model.Ad) error {
	if err := ad.Delete(); err != nil {
		return ucerrs.ErrCannotDelete
	}
	return uc.trManager.Do(ctx, func(txCtx context.Context) error {
		if err := uc.ad.Delete(txCtx, ad.ID()); err != nil {
			return ucerrs.Wrap(ucerrs.ErrDeleteAdDB, err)
		}
		if err := uc.media.Delete(txCtx, ad.ID()); err != nil {
			return ucerrs.Wrap(ucerrs.ErrDeleteImagesDB, err)
		}
		return nil
	})
}

// softDeleteAndPublish marks the ad as deleted, removes its media,
// and publishes an event so other services can react.
func (uc *DeleteAdUC) softDeleteAndPublish(ctx context.Context, ad *model.Ad) error {
	if err := ad.Delete(); err != nil {
		return ucerrs.ErrCannotDelete
	}
	return uc.trManager.Do(ctx, func(txCtx context.Context) error {
		if err := uc.ad.Update(txCtx, ad); err != nil {
			return ucerrs.Wrap(ucerrs.ErrUpdateAdDB, err)
		}
		if err := uc.media.Delete(txCtx, ad.ID()); err != nil {
			return ucerrs.Wrap(ucerrs.ErrDeleteImagesDB, err)
		}
		if err := uc.publisher.PublishAdDeleted(txCtx, ad.ID()); err != nil {
			return ucerrs.Wrap(ucerrs.ErrPublishAdDeleted, err)
		}
		return nil
	})
}
