package grpc

import (
	"github.com/maket12/ads-service/backend/authservice/api/proto/generated/auth_v2"
	"github.com/maket12/ads-service/backend/authservice/internal/app/dto"
	"github.com/maket12/ads-service/backend/authservice/pkg/utils"

	"github.com/google/uuid"
)

func MapRegisterPbToDTO(req *auth_v2.RegisterRequest) dto.RegisterInput {
	return dto.RegisterInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
}

func MapRegisterDTOToPb(out dto.RegisterOutput) *auth_v2.RegisterResponse {
	return &auth_v2.RegisterResponse{AccountId: out.AccountID.String()}
}

func MapLoginPbToDTO(req *auth_v2.LoginRequest) dto.LoginInput {
	return dto.LoginInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		IP:        utils.VPtr(req.GetIp()),
		UserAgent: utils.VPtr(req.GetUserAgent()),
	}
}

func MapLoginDTOToPb(out dto.LoginOutput) *auth_v2.LoginResponse {
	return &auth_v2.LoginResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func MapLogoutPbToDTO(req *auth_v2.LogoutRequest) dto.LogoutInput {
	return dto.LogoutInput{RefreshToken: req.GetRefreshToken()}
}

func MapLogoutDTOToPb(out dto.LogoutOutput) *auth_v2.LogoutResponse {
	return &auth_v2.LogoutResponse{Logout: out.Logout}
}

func MapRefreshSessionPbToDTO(req *auth_v2.RefreshSessionRequest) dto.RefreshSessionInput {
	return dto.RefreshSessionInput{
		RefreshToken: req.GetOldRefreshToken(),
		IP:           utils.VPtr(req.GetIp()),
		UserAgent:    utils.VPtr(req.GetUserAgent()),
	}
}

func MapRefreshSessionDTOToPb(out dto.RefreshSessionOutput) *auth_v2.RefreshSessionResponse {
	return &auth_v2.RefreshSessionResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
	}
}

func MapValidateAccessTokenPbToDTO(req *auth_v2.ValidateAccessTokenRequest) dto.ValidateAccessTokenInput {
	return dto.ValidateAccessTokenInput{AccessToken: req.GetAccessToken()}
}

func MapValidateAccessTokenDTOToPb(out dto.ValidateAccessTokenOutput) *auth_v2.ValidateAccessTokenResponse {
	return &auth_v2.ValidateAccessTokenResponse{
		AccountId: out.AccountID.String(),
		Role:      out.Role,
	}
}

func MapAssignRolePbToDTO(req *auth_v2.AssignRoleRequest) dto.AssignRoleInput {
	accID, _ := uuid.Parse(req.GetAccountId())
	return dto.AssignRoleInput{
		AccountID: accID,
		Role:      req.GetRole(),
	}
}

func MapAssignRoleDTOToPb(out dto.AssignRoleOutput) *auth_v2.AssignRoleResponse {
	return &auth_v2.AssignRoleResponse{Assigned: out.Assigned}
}

func MapSendVerificationPbToDTO(req *auth_v2.SendVerificationRequest) dto.SendVerificationInput {
	accID, _ := uuid.Parse(req.GetAccountId())
	return dto.SendVerificationInput{AccountID: accID}
}

func MapSendVerificationDTOToPb(out dto.SendVerificationOutput) *auth_v2.SendVerificationResponse {
	return &auth_v2.SendVerificationResponse{Sent: out.Sent}
}

func MapVerifyEmailPbToDTO(req *auth_v2.VerifyEmailRequest) dto.VerifyEmailInput {
	return dto.VerifyEmailInput{Token: req.GetToken()}
}

func MapVerifyEmailDTOToPb(out dto.VerifyEmailOutput) *auth_v2.VerifyEmailResponse {
	return &auth_v2.VerifyEmailResponse{Verified: out.Verified}
}

func MapBlockAccountPbToDTO(req *auth_v2.BlockAccountRequest) dto.BlockAccountInput {
	accID, _ := uuid.Parse(req.GetAccountId())
	return dto.BlockAccountInput{AccountID: accID}
}

func MapBlockAccountDTOToPb(out dto.BlockAccountOutput) *auth_v2.BlockAccountResponse {
	return &auth_v2.BlockAccountResponse{Blocked: out.Blocked}
}

func MapDeleteAccountPbToDTO(req *auth_v2.DeleteAccountRequest) dto.DeleteAccountInput {
	accID, _ := uuid.Parse(req.GetAccountId())
	return dto.DeleteAccountInput{AccountID: accID}
}

func MapDeleteAccountDTOToPb(out dto.DeleteAccountOutput) *auth_v2.DeleteAccountResponse {
	return &auth_v2.DeleteAccountResponse{Deleted: out.Deleted}
}
