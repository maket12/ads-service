package grpc

import (
	"ads/adservice/internal/app/usecase"
	"ads/pkg/generated/ad_v1"
	"ads/pkg/utils"
	"context"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type AdHandler struct {
	ad_v1.UnimplementedAdServiceServer
	log            *slog.Logger
	createAdUC     *usecase.CreateAdUC
	getAdUC        *usecase.GetAdUC
	updateAdUC     *usecase.UpdateAdUC
	publishAdUC    *usecase.PublishAdUC
	rejectAdUC     *usecase.RejectAdUC
	deleteAdUC     *usecase.DeleteAdUC
	deleteAllAdsUC *usecase.DeleteAllAdsUC
}

func NewAdHandler(
	log *slog.Logger,
	createAdUC *usecase.CreateAdUC,
	getAdUC *usecase.GetAdUC,
	updateAdUC *usecase.UpdateAdUC,
	publishAdUC *usecase.PublishAdUC,
	rejectAdUC *usecase.RejectAdUC,
	deleteAdUC *usecase.DeleteAdUC,
	deleteAllAdsUC *usecase.DeleteAllAdsUC,
) *AdHandler {
	return &AdHandler{
		log:            log,
		createAdUC:     createAdUC,
		getAdUC:        getAdUC,
		updateAdUC:     updateAdUC,
		publishAdUC:    publishAdUC,
		rejectAdUC:     rejectAdUC,
		deleteAdUC:     deleteAdUC,
		deleteAllAdsUC: deleteAllAdsUC,
	}
}

// Extracts account id from context and returns gRPC error if fails
func (h *AdHandler) extractID(ctx context.Context) (uuid.UUID, error) {
	accountID, err := utils.ExtractAccountID(ctx)
	if err != nil {
		outErr := gRPCError(err)
		return uuid.Nil, status.Error(outErr.Code, outErr.Message)
	}
	return accountID, nil
}

func (h *AdHandler) CreateAd(ctx context.Context, req *ad_v1.CreateAdRequest) (*ad_v1.CreateAdResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.createAdUC.Execute(ctx, MapCreateAdPbToDTO(req, accountID))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to create ad",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapCreateAdDTOToPb(ucResp), nil
}

func (h *AdHandler) GetAd(ctx context.Context, req *ad_v1.GetAdRequest) (*ad_v1.GetAdResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.getAdUC.Execute(ctx, MapGetAdPbToDTO(req, accountID))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to get ad",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapGetAdDTOToPb(ucResp), nil
}

func (h *AdHandler) UpdateAd(ctx context.Context, req *ad_v1.UpdateAdRequest) (*ad_v1.UpdateAdResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.updateAdUC.Execute(ctx, MapUpdateAdPbToDTO(req, accountID))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to update ad",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapUpdateAdDTOToPb(ucResp), nil
}

func (h *AdHandler) PublishAd(ctx context.Context, req *ad_v1.PublishAdRequest) (*ad_v1.PublishAdResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.publishAdUC.Execute(ctx, MapPublishAdPbToDTO(req, accountID))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to publish ad",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapPublishAdDTOToPb(ucResp), nil
}

func (h *AdHandler) RejectAd(ctx context.Context, req *ad_v1.RejectAdRequest) (*ad_v1.RejectAdResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.rejectAdUC.Execute(ctx, MapRejectAdPbToDTO(req, accountID))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to reject ad",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapRejectAdDTOToPb(ucResp), nil
}

func (h *AdHandler) DeleteAd(ctx context.Context, req *ad_v1.DeleteAdRequest) (*ad_v1.DeleteAdResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.deleteAdUC.Execute(ctx, MapDeleteAdPbToDTO(req, accountID))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to delete ad",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapDeleteAdDTOToPb(ucResp), nil
}

func (h *AdHandler) DeleteAllAds(ctx context.Context, req *ad_v1.DeleteAllAdsRequest) (*ad_v1.DeleteAllAdsResponse, error) {
	ucResp, err := h.deleteAllAdsUC.Execute(ctx, MapDeleteAllAdsPbToDTO(req))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to delete all ads",
			slog.String("seller_id", req.GetSellerId()),
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapDeleteAllAdsDTOToPb(ucResp), nil
}
