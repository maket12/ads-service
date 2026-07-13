package grpc

import (
	"context"
	"log/slog"

	"github.com/maket12/ads-service/pkg/generated/user_v1"
	"github.com/maket12/ads-service/pkg/utils"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	"github.com/maket12/ads-service/userservice/internal/app/usecase"

	"github.com/google/uuid"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	user_v1.UnimplementedUserServiceServer
	log             *slog.Logger
	getProfileUC    *usecase.GetProfileUC
	updateProfileUC *usecase.UpdateProfileUC
}

func NewUserHandler(
	log *slog.Logger,
	getProfileUC *usecase.GetProfileUC,
	updateProfileUC *usecase.UpdateProfileUC,
) *UserHandler {
	return &UserHandler{
		log:             log,
		getProfileUC:    getProfileUC,
		updateProfileUC: updateProfileUC,
	}
}

// Extracts account id from context and returns gRPC error if fails
func (h *UserHandler) extractID(ctx context.Context) (uuid.UUID, error) {
	accountID, err := utils.ExtractAccountID(ctx)
	if err != nil {
		outErr := gRPCError(err)
		return uuid.Nil, status.Error(outErr.Code, outErr.Message)
	}
	return accountID, nil
}

func (h *UserHandler) GetProfile(ctx context.Context, _ *user_v1.GetProfileRequest) (*user_v1.GetProfileResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.getProfileUC.Execute(
		ctx, dto.GetProfileInput{AccountID: accountID},
	)

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to get profile",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapGetProfileDTOToPb(ucResp), nil
}

func (h *UserHandler) UpdateProfile(ctx context.Context, req *user_v1.UpdateProfileRequest) (*user_v1.UpdateProfileResponse, error) {
	accountID, gRPCErr := h.extractID(ctx)
	if gRPCErr != nil {
		return nil, gRPCErr
	}

	ucResp, err := h.updateProfileUC.Execute(ctx,
		MapUpdateProfilePbToDTO(accountID, req),
	)

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to update profile",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapUpdateProfileDTOToPb(ucResp), nil
}
