package grpc

import (
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	"github.com/maket12/ads-service/pkg/generated/auth_v1"

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
	var ip, userAgent = req.GetIp(), req.GetUserAgent()
	return dto.LoginInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		IP:        &ip,
		UserAgent: &userAgent,
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
	var ip, userAgent = req.GetIp(), req.GetUserAgent()
	return dto.RefreshSessionInput{
		OldRefreshToken: req.GetOldRefreshToken(),
		IP:              &ip,
		UserAgent:       &userAgent,
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
	return &auth_v1.AssignRoleResponse{Assign: out.Assign}
}
