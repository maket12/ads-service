package grpc

import (
	"context"
	"log/slog"

	"github.com/maket12/ads-service/authservice/internal/app/usecase"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	auth_v1.UnimplementedAuthServiceServer
	log                   *slog.Logger
	registerUC            *usecase.RegisterUC
	loginUC               *usecase.LoginUC
	logoutUC              *usecase.LogoutUC
	refreshSessionUC      *usecase.RefreshSessionUC
	validateAccessTokenUC *usecase.ValidateAccessTokenUC
	assignRoleUC          *usecase.AssignRoleUC
	sendVerificationUC    *usecase.SendVerificationUC
	verifyEmailUC         *usecase.VerifyEmailUC
}

func NewAuthHandler(
	log *slog.Logger,
	registerUC *usecase.RegisterUC,
	loginUC *usecase.LoginUC,
	logoutUC *usecase.LogoutUC,
	refreshSessionUC *usecase.RefreshSessionUC,
	validateAccessTokenUC *usecase.ValidateAccessTokenUC,
	assignRoleUC *usecase.AssignRoleUC,
	sendVerificationUC *usecase.SendVerificationUC,
	verifyEmailUC *usecase.VerifyEmailUC,
) *AuthHandler {
	return &AuthHandler{
		log:                   log,
		registerUC:            registerUC,
		loginUC:               loginUC,
		logoutUC:              logoutUC,
		refreshSessionUC:      refreshSessionUC,
		validateAccessTokenUC: validateAccessTokenUC,
		assignRoleUC:          assignRoleUC,
		sendVerificationUC:    sendVerificationUC,
		verifyEmailUC:         verifyEmailUC,
	}
}

func (h *AuthHandler) Register(
	ctx context.Context,
	req *auth_v1.RegisterRequest,
) (*auth_v1.RegisterResponse, error) {
	ucResp, err := h.registerUC.Execute(ctx, MapRegisterPbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to register")
		return nil, status.Error(code, msg)
	}

	h.log.InfoContext(ctx, "created account",
		slog.String("email", req.GetEmail()),
	)

	return MapRegisterDTOToPb(ucResp), nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	req *auth_v1.LoginRequest,
) (*auth_v1.LoginResponse, error) {
	ucResp, err := h.loginUC.Execute(ctx, MapLoginPbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to login")
		return nil, status.Error(code, msg)
	}

	h.log.InfoContext(ctx, "successful login",
		slog.String("email", req.GetEmail()),
		slog.String("ip", req.GetIp()),
		slog.String("user_agent", req.GetUserAgent()),
	)

	return MapLoginDTOToPb(ucResp), nil
}

func (h *AuthHandler) Logout(
	ctx context.Context,
	req *auth_v1.LogoutRequest,
) (*auth_v1.LogoutResponse, error) {
	ucResp, err := h.logoutUC.Execute(ctx, MapLogoutPbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to logout")
		return nil, status.Error(code, msg)
	}

	return MapLogoutDTOToPb(ucResp), nil
}

func (h *AuthHandler) RefreshSession(
	ctx context.Context,
	req *auth_v1.RefreshSessionRequest,
) (*auth_v1.RefreshSessionResponse, error) {
	ucResp, err := h.refreshSessionUC.Execute(ctx, MapRefreshSessionPbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to refresh session")
		return nil, status.Error(code, msg)
	}

	return MapRefreshSessionDTOToPb(ucResp), nil
}

func (h *AuthHandler) ValidateAccessToken(
	ctx context.Context,
	req *auth_v1.ValidateAccessTokenRequest,
) (*auth_v1.ValidateAccessTokenResponse, error) {
	ucResp, err := h.validateAccessTokenUC.Execute(ctx, MapValidateAccessTokenPbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to validate access token")
		return nil, status.Error(code, msg)
	}

	return MapValidateAccessTokenDTOToPb(ucResp), nil
}

func (h *AuthHandler) AssignRole(
	ctx context.Context,
	req *auth_v1.AssignRoleRequest,
) (*auth_v1.AssignRoleResponse, error) {
	ucResp, err := h.assignRoleUC.Execute(ctx, MapAssignRolePbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to assign account role")
		return nil, status.Error(code, msg)
	}

	h.log.InfoContext(ctx, "successful role assigned",
		slog.String("account_id", req.GetAccountId()),
		slog.String("role", req.GetRole()),
	)

	return MapAssignRoleDTOToPb(ucResp), nil
}

func (h *AuthHandler) SendVerification(
	ctx context.Context,
	req *auth_v1.SendVerificationRequest,
) (*auth_v1.SendVerificationResponse, error) {
	ucResp, err := h.sendVerificationUC.Execute(ctx, MapSendVerificationPbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to send verification")
		return nil, status.Error(code, msg)
	}

	h.log.InfoContext(ctx, "sent verification token",
		slog.String("account_id", req.GetAccountId()),
	)

	return MapSendVerificationDTOToPb(ucResp), nil
}

func (h *AuthHandler) VerifyEmail(
	ctx context.Context,
	req *auth_v1.VerifyEmailRequest,
) (*auth_v1.VerifyEmailResponse, error) {
	ucResp, err := h.verifyEmailUC.Execute(ctx, MapVerifyEmailPbToDTO(req))

	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to verify email")
		return nil, status.Error(code, msg)
	}

	return MapVerifyEmailDTOToPb(ucResp), nil
}

func (h *AuthHandler) handleError(
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
