package grpc

import (
	"ads/pkg/generated/user_v1"
	"ads/userservice/internal/app/dto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapGetProfileDTOToPb(out dto.GetProfileOutput) *user_v1.GetProfileResponse {
	return &user_v1.GetProfileResponse{
		AccountId: out.AccountID.String(),
		FirstName: out.FirstName,
		LastName:  out.LastName,
		Phone:     out.Phone,
		AvatarUrl: out.AvatarURL,
		Bio:       out.Bio,
		UpdatedAt: timestamppb.New(out.UpdatedAt),
	}
}

func MapUpdateProfilePbToDTO(accountID uuid.UUID, req *user_v1.UpdateProfileRequest) dto.UpdateProfileInput {
	return dto.UpdateProfileInput{
		AccountID: accountID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		AvatarURL: req.AvatarUrl,
		Bio:       req.Bio,
	}
}

func MapUpdateProfileDTOToPb(out dto.UpdateProfileOutput) *user_v1.UpdateProfileResponse {
	return &user_v1.UpdateProfileResponse{Success: out.Success}
}
