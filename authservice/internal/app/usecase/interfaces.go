package usecase

import (
	"context"

	"github.com/maket12/ads-service/authservice/internal/app/dto"
)

type AssignRoleUseCase interface {
	Execute(ctx context.Context, in dto.AssignRoleInput) (dto.AssignRoleOutput, error)
}

type LoginUseCase interface {
	Execute(ctx context.Context, in dto.LoginInput) (dto.LoginOutput, error)
}

type LogoutUseCase interface {
	Execute(ctx context.Context, in dto.LogoutInput) (dto.LogoutOutput, error)
}

type RefreshSessionUseCase interface {
	Execute(ctx context.Context, in dto.RefreshSessionInput) (dto.RefreshSessionOutput, error)
}

type RegisterUseCase interface {
	Execute(ctx context.Context, in dto.RegisterInput) (dto.RegisterOutput, error)
}

type ValidateAccessTokenUseCase interface {
	Execute(ctx context.Context, in dto.ValidateAccessTokenInput) (dto.ValidateAccessTokenOutput, error)
}
