package grpc

import (
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/app/dto"
	"github.com/maket12/ads-service/authservice/pkg/generated/auth_v1"
	"github.com/maket12/ads-service/authservice/pkg/utils"
	"github.com/stretchr/testify/require"
)

func TestMapRegisterPbToDTO(t *testing.T) {
	email := gofakeit.Email()
	pass := gofakeit.Password(true, true, true, true, true, 10)

	req := &auth_v1.RegisterRequest{Email: email, Password: pass}
	expected := dto.RegisterInput{Email: email, Password: pass}
	actual := MapRegisterPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapRegisterDTOToPb(t *testing.T) {
	accID := uuid.New()

	out := dto.RegisterOutput{AccountID: accID}
	expected := &auth_v1.RegisterResponse{AccountId: accID.String()}
	actual := MapRegisterDTOToPb(out)

	require.Equal(t, expected, actual)
}

func TestMapLoginPbToDTO(t *testing.T) {
	email := gofakeit.Email()
	pass := gofakeit.Password(true, true, true, true, true, 10)
	ip := gofakeit.IPv4Address()
	ua := gofakeit.UserAgent()

	req := &auth_v1.LoginRequest{
		Email:     email,
		Password:  pass,
		Ip:        &ip,
		UserAgent: &ua,
	}

	expected := dto.LoginInput{
		Email:     email,
		Password:  pass,
		IP:        utils.VPtr(ip),
		UserAgent: utils.VPtr(ua),
	}

	actual := MapLoginPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapLoginDTOToPb(t *testing.T) {
	accessToken := gofakeit.UUID()
	refreshToken := gofakeit.UUID()

	out := dto.LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	expected := &auth_v1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	actual := MapLoginDTOToPb(out)

	require.Equal(t, expected, actual)
}

func TestMapLogoutPbToDTO(t *testing.T) {
	refreshToken := gofakeit.UUID()

	req := &auth_v1.LogoutRequest{RefreshToken: refreshToken}
	expected := dto.LogoutInput{RefreshToken: refreshToken}
	actual := MapLogoutPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapLogoutDTOToPb(t *testing.T) {
	logout := gofakeit.Bool()

	out := dto.LogoutOutput{Logout: logout}
	expected := &auth_v1.LogoutResponse{Logout: logout}
	actual := MapLogoutDTOToPb(out)

	require.Equal(t, expected, actual)
}

func TestMapRefreshSessionPbToDTO(t *testing.T) {
	oldRefreshToken := gofakeit.UUID()
	ip := gofakeit.IPv4Address()
	ua := gofakeit.UserAgent()

	req := &auth_v1.RefreshSessionRequest{
		OldRefreshToken: oldRefreshToken,
		Ip:              &ip,
		UserAgent:       &ua,
	}

	expected := dto.RefreshSessionInput{
		RefreshToken: oldRefreshToken,
		IP:           utils.VPtr(ip),
		UserAgent:    utils.VPtr(ua),
	}

	actual := MapRefreshSessionPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapRefreshSessionDTOToPb(t *testing.T) {
	accessToken := gofakeit.UUID()
	refreshToken := gofakeit.UUID()

	out := dto.RefreshSessionOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	expected := &auth_v1.RefreshSessionResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	actual := MapRefreshSessionDTOToPb(out)

	require.Equal(t, expected, actual)
}

func TestMapValidateAccessTokenPbToDTO(t *testing.T) {
	accessToken := gofakeit.UUID()

	req := &auth_v1.ValidateAccessTokenRequest{AccessToken: accessToken}
	expected := dto.ValidateAccessTokenInput{AccessToken: accessToken}
	actual := MapValidateAccessTokenPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapValidateAccessTokenDTOToPb(t *testing.T) {
	accID := uuid.New()
	role := gofakeit.RandomString([]string{"admin", "user"})

	out := dto.ValidateAccessTokenOutput{
		AccountID: accID,
		Role:      role,
	}

	expected := &auth_v1.ValidateAccessTokenResponse{
		AccountId: accID.String(),
		Role:      role,
	}

	actual := MapValidateAccessTokenDTOToPb(out)

	require.Equal(t, expected, actual)
}

func TestMapAssignRolePbToDTO(t *testing.T) {
	accID := uuid.New()
	role := gofakeit.RandomString([]string{"admin", "user"})

	req := &auth_v1.AssignRoleRequest{
		AccountId: accID.String(),
		Role:      role,
	}

	expected := dto.AssignRoleInput{
		AccountID: accID,
		Role:      role,
	}

	actual := MapAssignRolePbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapAssignRolePbToDTO_InvalidAccountID(t *testing.T) {
	role := gofakeit.RandomString([]string{"admin", "user"})

	req := &auth_v1.AssignRoleRequest{
		AccountId: "not-a-valid-uuid",
		Role:      role,
	}

	expected := dto.AssignRoleInput{
		AccountID: uuid.UUID{},
		Role:      role,
	}

	actual := MapAssignRolePbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapAssignRoleDTOToPb(t *testing.T) {
	assigned := gofakeit.Bool()

	out := dto.AssignRoleOutput{Assigned: assigned}
	expected := &auth_v1.AssignRoleResponse{Assigned: assigned}
	actual := MapAssignRoleDTOToPb(out)

	require.Equal(t, expected, actual)
}

func TestMapSendVerificationPbToDTO(t *testing.T) {
	accID := uuid.New()

	req := &auth_v1.SendVerificationRequest{AccountId: accID.String()}
	expected := dto.SendVerificationInput{AccountID: accID}
	actual := MapSendVerificationPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapSendVerificationPbToDTO_InvalidAccountID(t *testing.T) {
	req := &auth_v1.SendVerificationRequest{AccountId: "not-a-valid-uuid"}
	expected := dto.SendVerificationInput{AccountID: uuid.UUID{}}
	actual := MapSendVerificationPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapSendVerificationDTOToPb(t *testing.T) {
	sent := gofakeit.Bool()

	out := dto.SendVerificationOutput{Sent: sent}
	expected := &auth_v1.SendVerificationResponse{Sent: sent}
	actual := MapSendVerificationDTOToPb(out)

	require.Equal(t, expected, actual)
}

func TestMapVerifyEmailPbToDTO(t *testing.T) {
	token := gofakeit.UUID()

	req := &auth_v1.VerifyEmailRequest{Token: token}
	expected := dto.VerifyEmailInput{Token: token}
	actual := MapVerifyEmailPbToDTO(req)

	require.Equal(t, expected, actual)
}

func TestMapVerifyEmailDTOToPb(t *testing.T) {
	verified := gofakeit.Bool()

	out := dto.VerifyEmailOutput{Verified: verified}
	expected := &auth_v1.VerifyEmailResponse{Verified: verified}
	actual := MapVerifyEmailDTOToPb(out)

	require.Equal(t, expected, actual)
}
