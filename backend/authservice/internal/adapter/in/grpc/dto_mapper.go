package grpc

import (
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	"github.com/maket12/ads-service/authservice/pkg/utils"

	"github.com/google/uuid"
)

func MapRegisterPbToDTO(req *auth_v1.RegisterRequest) dto.RegisterInput {
	return dto.RegisterInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
}

func MapRegisterDTOToPb(out dto.RegisterOutput) *auth_v1.RegisterResponse {
	return &auth_v1.RegisterResponse{AccountId: out.AccountID.String()}
}

func MapLoginPbToDTO(req *auth_v1.LoginRequest) dto.LoginInput {
	return dto.LoginInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		IP:        utils.VPtr(req.GetIp()),
		UserAgent: utils.VPtr(req.GetUserAgent()),
	}
}

func MapLoginDTOToPb(out dto.LoginOutput) *auth_v1.LoginResponse {
	return &auth_v1.LoginResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func MapLogoutPbToDTO(req *auth_v1.LogoutRequest) dto.LogoutInput {
	return dto.LogoutInput{RefreshToken: req.GetRefreshToken()}
}

func MapLogoutDTOToPb(out dto.LogoutOutput) *auth_v1.LogoutResponse {
	return &auth_v1.LogoutResponse{Logout: out.Logout}
}

func MapRefreshSessionPbToDTO(req *auth_v1.RefreshSessionRequest) dto.RefreshSessionInput {
	return dto.RefreshSessionInput{
		RefreshToken: req.GetOldRefreshToken(),
		IP:           utils.VPtr(req.GetIp()),
		UserAgent:    utils.VPtr(req.GetUserAgent()),
	}
}

func MapRefreshSessionDTOToPb(out dto.RefreshSessionOutput) *auth_v1.RefreshSessionResponse {
	return &auth_v1.RefreshSessionResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func MapValidateAccessTokenPbToDTO(req *auth_v1.ValidateAccessTokenRequest) dto.ValidateAccessTokenInput {
	return dto.ValidateAccessTokenInput{AccessToken: req.GetAccessToken()}
}

func MapValidateAccessTokenDTOToPb(out dto.ValidateAccessTokenOutput) *auth_v1.ValidateAccessTokenResponse {
	return &auth_v1.ValidateAccessTokenResponse{
		AccountId: out.AccountID.String(),
		Role:      out.Role,
	}
}

func MapAssignRolePbToDTO(req *auth_v1.AssignRoleRequest) dto.AssignRoleInput {
	accID, _ := uuid.Parse(req.GetAccountId())
	return dto.AssignRoleInput{
		AccountID: accID,
		Role:      req.GetRole(),
	}
}

func MapAssignRoleDTOToPb(out dto.AssignRoleOutput) *auth_v1.AssignRoleResponse {
	return &auth_v1.AssignRoleResponse{Assigned: out.Assigned}
}

func MapSendVerificationPbToDTO(req *auth_v1.SendVerificationRequest) dto.SendVerificationInput {
	accID, _ := uuid.Parse(req.GetAccountId())
	return dto.SendVerificationInput{
		AccountID: accID,
		Email:     req.GetEmail(),
	}
}

func MapSendVerificationDTOToPb(out dto.SendVerificationOutput) *auth_v1.SendVerificationResponse {
	return &auth_v1.SendVerificationResponse{Sent: out.Sent}
}

func MapVerifyEmailPbToDTO(req *auth_v1.VerifyEmailRequest) dto.VerifyEmailInput {
	return dto.VerifyEmailInput{Token: req.GetToken()}
}

func MapVerifyEmailDTOToPb(out dto.VerifyEmailOutput) *auth_v1.VerifyEmailResponse {
	return &auth_v1.VerifyEmailResponse{Verified: out.Verified}
}
