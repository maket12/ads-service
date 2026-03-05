package grpc

import (
	"context"
	"log/slog"

	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/pkg/generated/auth_v1"

	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	auth_v1.UnimplementedAuthServiceServer
	log                   *slog.Logger
	registerUC            usecase.RegisterUseCase
	loginUC               usecase.LoginUseCase
	logoutUC              usecase.LogoutUseCase
	refreshSessionUC      usecase.RefreshSessionUseCase
	validateAccessTokenUC usecase.ValidateAccessTokenUseCase
	assignRoleUC          usecase.AssignRoleUseCase
}

func NewAuthHandler(
	log *slog.Logger,
	registerUC usecase.RegisterUseCase,
	loginUC usecase.LoginUseCase,
	logoutUC usecase.LogoutUseCase,
	refreshSessionUC usecase.RefreshSessionUseCase,
	validateAccessTokenUC usecase.ValidateAccessTokenUseCase,
	assignRoleUC usecase.AssignRoleUseCase,
) *AuthHandler {
	return &AuthHandler{
		log:                   log,
		registerUC:            registerUC,
		loginUC:               loginUC,
		logoutUC:              logoutUC,
		refreshSessionUC:      refreshSessionUC,
		validateAccessTokenUC: validateAccessTokenUC,
		assignRoleUC:          assignRoleUC,
	}
}

func (h *AuthHandler) Register(ctx context.Context, req *auth_v1.RegisterRequest) (*auth_v1.RegisterResponse, error) {
	ucResp, err := h.registerUC.Execute(ctx, MapRegisterPbToDTO(req))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to register",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapRegisterDTOToPb(ucResp), nil
}

func (h *AuthHandler) Login(ctx context.Context, req *auth_v1.LoginRequest) (*auth_v1.LoginResponse, error) {
	ucResp, err := h.loginUC.Execute(ctx, MapLoginPbToDTO(req))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to login",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapLoginDTOToPb(ucResp), nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *auth_v1.LogoutRequest) (*auth_v1.LogoutResponse, error) {
	ucResp, err := h.logoutUC.Execute(ctx, MapLogoutPbToDTO(req))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to logout",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapLogoutDTOToPb(ucResp), nil
}

func (h *AuthHandler) RefreshSession(ctx context.Context, req *auth_v1.RefreshSessionRequest) (*auth_v1.RefreshSessionResponse, error) {
	ucResp, err := h.refreshSessionUC.Execute(ctx, MapRefreshSessionPbToDTO(req))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to refresh session",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapRefreshSessionDTOToPb(ucResp), nil
}

func (h *AuthHandler) ValidateAccessToken(ctx context.Context, req *auth_v1.ValidateAccessTokenRequest) (*auth_v1.ValidateAccessTokenResponse, error) {
	ucResp, err := h.validateAccessTokenUC.Execute(ctx, MapValidateAccessTokenPbToDTO(req))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to validate access token",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapValidateAccessTokenDTOToPb(ucResp), nil
}

func (h *AuthHandler) AssignRole(ctx context.Context, req *auth_v1.AssignRoleRequest) (*auth_v1.AssignRoleResponse, error) {
	ucResp, err := h.assignRoleUC.Execute(ctx, MapAssignRolePbToDTO(req))

	if err != nil {
		outErr := gRPCError(err)
		h.log.ErrorContext(ctx, "failed to assign account role",
			slog.Int("code", int(outErr.Code)),
			slog.String("public_msg", outErr.Message),
			slog.Any("reason", outErr.Reason),
		)
		return nil, status.Error(outErr.Code, outErr.Message)
	}

	return MapAssignRoleDTOToPb(ucResp), nil
}
