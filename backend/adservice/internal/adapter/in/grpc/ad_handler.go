package grpc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/maket12/ads-service/backend/adservice/v2/internal/app/usecase"
	"github.com/maket12/ads-service/backend/adservice/v2/pkg/generated/ad_v1"
	"github.com/maket12/ads-service/backend/adservice/v2/pkg/utils"
	"google.golang.org/grpc/codes"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type AdHandler struct {
	ad_v1.UnimplementedAdServiceServer
	log             *slog.Logger
	createAdUC      *usecase.CreateAdUC
	getAdUC         *usecase.GetAdUC
	updateAdUC      *usecase.UpdateAdUC
	publishAdUC     *usecase.PublishAdUC
	rejectAdUC      *usecase.RejectAdUC
	deleteAdUC      *usecase.DeleteAdUC
	deleteAllAdsUC  *usecase.DeleteAllAdsUC
	listSellerAdsUC *usecase.ListSellerAdsUC
	listAllAdsUC    *usecase.ListAllAdsUC
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
	listSellerAdsUC *usecase.ListSellerAdsUC,
	listAllAdsUC *usecase.ListAllAdsUC,
) *AdHandler {
	return &AdHandler{
		log:             log,
		createAdUC:      createAdUC,
		getAdUC:         getAdUC,
		updateAdUC:      updateAdUC,
		publishAdUC:     publishAdUC,
		rejectAdUC:      rejectAdUC,
		deleteAdUC:      deleteAdUC,
		deleteAllAdsUC:  deleteAllAdsUC,
		listSellerAdsUC: listSellerAdsUC,
		listAllAdsUC:    listAllAdsUC,
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

// Extracts account role from context and returns gRPC error if fails
func (h *AdHandler) extractRole(ctx context.Context) (string, error) {
	role, err := utils.ExtractAccountRole(ctx)
	if err != nil {
		outErr := gRPCError(err)
		return "", status.Error(outErr.Code, outErr.Message)
	}
	return role, nil
}

func (h *AdHandler) CreateAd(ctx context.Context, req *ad_v1.CreateAdRequest) (*ad_v1.CreateAdResponse, error) {
	accountID, err := h.authorize(ctx, "create-ad", false)
	if err != nil {
		return nil, err
	}

	ucResp, err := h.createAdUC.Execute(ctx, MapCreateAdPbToDTO(req, accountID))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to create ad")
		return nil, status.Error(code, msg)
	}

	h.log.InfoContext(ctx, "created ad",
		slog.String("title", req.GetTitle()),
		slog.String("description", req.GetDescription()),
		slog.Int("price", int(req.GetPrice())),
	)

	return MapCreateAdDTOToPb(ucResp), nil
}

func (h *AdHandler) GetAd(ctx context.Context, req *ad_v1.GetAdRequest) (*ad_v1.GetAdResponse, error) {
	accountID, err := h.authorize(ctx, "get-ad", false)
	if err != nil {
		return nil, err
	}

	ucResp, err := h.getAdUC.Execute(ctx, MapGetAdPbToDTO(req, accountID))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to get ad")
		return nil, status.Error(code, msg)
	}

	return MapGetAdDTOToPb(ucResp), nil
}

func (h *AdHandler) UpdateAd(ctx context.Context, req *ad_v1.UpdateAdRequest) (*ad_v1.UpdateAdResponse, error) {
	accountID, err := h.authorize(ctx, "update-ad", false)
	if err != nil {
		return nil, err
	}

	ucResp, err := h.updateAdUC.Execute(ctx, MapUpdateAdPbToDTO(req, accountID))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to update ad")
		return nil, status.Error(code, msg)
	}

	h.log.InfoContext(ctx, "updated ad",
		slog.String("id", req.GetId()),
		slog.String("description", req.GetDescription()),
		slog.String("title", req.GetTitle()),
		slog.Int("price", int(req.GetPrice())),
	)

	return MapUpdateAdDTOToPb(ucResp), nil
}

func (h *AdHandler) PublishAd(ctx context.Context, req *ad_v1.PublishAdRequest) (*ad_v1.PublishAdResponse, error) {
	if _, err := h.authorize(ctx, "publish-ad", true); err != nil {
		return nil, err
	}

	ucResp, err := h.publishAdUC.Execute(ctx, MapPublishAdPbToDTO(req))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to publish ad")
		return nil, status.Error(code, msg)
	}

	slog.InfoContext(ctx, "published ad",
		slog.String("id", req.GetId()),
	)

	return MapPublishAdDTOToPb(ucResp), nil
}

func (h *AdHandler) RejectAd(ctx context.Context, req *ad_v1.RejectAdRequest) (*ad_v1.RejectAdResponse, error) {
	if _, err := h.authorize(ctx, "reject-ad", true); err != nil {
		return nil, err
	}

	ucResp, err := h.rejectAdUC.Execute(ctx, MapRejectAdPbToDTO(req))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to reject ad")
		return nil, status.Error(code, msg)
	}

	slog.InfoContext(ctx, "rejected ad",
		slog.String("id", req.GetId()),
	)

	return MapRejectAdDTOToPb(ucResp), nil
}

func (h *AdHandler) DeleteAd(ctx context.Context, req *ad_v1.DeleteAdRequest) (*ad_v1.DeleteAdResponse, error) {
	accountID, err := h.authorize(ctx, "reject-ad", false)
	if err != nil {
		return nil, err
	}

	ucResp, err := h.deleteAdUC.Execute(ctx, MapDeleteAdPbToDTO(req, accountID))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to delete ad")
		return nil, status.Error(code, msg)
	}

	slog.InfoContext(ctx, "deleted ad",
		slog.String("id", req.GetId()),
	)

	return MapDeleteAdDTOToPb(ucResp), nil
}

func (h *AdHandler) DeleteAllAds(ctx context.Context, req *ad_v1.DeleteAllAdsRequest) (*ad_v1.DeleteAllAdsResponse, error) {
	if _, err := h.authorize(ctx, "delete-all-ads", true); err != nil {
		return nil, err
	}

	ucResp, err := h.deleteAllAdsUC.Execute(ctx, MapDeleteAllAdsPbToDTO(req))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to delete all ads")
		return nil, status.Error(code, msg)
	}

	slog.InfoContext(ctx, "deleted all ads")

	return MapDeleteAllAdsDTOToPb(ucResp), nil
}

func (h *AdHandler) ListAds(ctx context.Context, req *ad_v1.ListAdsRequest) (*ad_v1.ListAdsResponse, error) {
	accountID, err := h.authorize(ctx, "list-ads", false)
	if err != nil {
		return nil, err
	}

	ucResp, err := h.listSellerAdsUC.Execute(ctx, MapListAdsPbToDTO(req, accountID))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to get a list of ads")
		return nil, status.Error(code, msg)
	}

	return MapListAdsDTOToPb(ucResp), nil
}

func (h *AdHandler) ListAllAds(ctx context.Context, req *ad_v1.ListAllAdsRequest) (*ad_v1.ListAllAdsResponse, error) {
	if _, err := h.authorize(ctx, "list-all-ads", true); err != nil {
		return nil, err
	}

	ucResp, err := h.listAllAdsUC.Execute(ctx, MapListAllAdsPbToDTO(req))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to get a list of ads for admin")
		return nil, status.Error(code, msg)
	}

	return MapListAllAdsDTOToPb(ucResp), nil
}

func (h *AdHandler) authorize(
	ctx context.Context,
	method string, admin bool,
) (uuid.UUID, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return uuid.Nil, gRPCErr
	}

	if !admin {
		return accountID, nil
	}

	role, gRPCErr := h.extractRole(ctx)
	if gRPCErr != nil {
		return uuid.Nil, gRPCErr
	}

	if role != "admin" {
		methodName := fmt.Sprintf("[%s]", method)
		h.log.WarnContext(ctx, methodName+" access denied for non-admin user",
			slog.String("account_id", accountID.String()),
			slog.String("role", role),
		)
		return uuid.Nil, status.Error(codes.PermissionDenied, "access denied: admin role required")
	}

	return accountID, nil
}

func (h *AdHandler) handleError(
	ctx context.Context, err error,
	logMsg string,
) (codes.Code, string) {
	outErr := gRPCError(err)
	h.log.LogAttrs(ctx, outErr.Level, logMsg,
		slog.Int("code", int(outErr.Code)),
		slog.String("public_msg", outErr.Message),
		slog.Any("reason", outErr.Reason),
	)
	return outErr.Code, outErr.Message
}
