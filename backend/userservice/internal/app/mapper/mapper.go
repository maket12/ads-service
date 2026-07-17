package mapper

import (
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
)

func MapProfileToGetProfileDTO(p *model.Profile) dto.GetProfileOutput {
	return dto.GetProfileOutput{
		AccountID: p.AccountID(),
		FirstName: p.FirstName(),
		LastName:  p.LastName(),
		Phone:     p.Phone(),
		AvatarURL: p.AvatarURL(),
		Bio:       p.Bio(),
		UpdatedAt: p.UpdatedAt(),
	}
}
