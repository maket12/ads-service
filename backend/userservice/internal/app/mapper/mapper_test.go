package mapper_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/userservice/internal/app/mapper"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"github.com/maket12/ads-service/userservice/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMapProfileToGetProfileDTO(t *testing.T) {
	profile := model.RestoreProfile(uuid.New(),
		utils.VPtr(gofakeit.FirstName()),
		utils.VPtr(gofakeit.LastName()),
		utils.VPtr(gofakeit.PhoneFormatted()),
		utils.VPtr("https://web3-ui.com/cyberpunk-view.png"),
		utils.VPtr("Not a bad person"), time.Now(),
	)

	mapped := mapper.MapProfileToGetProfileDTO(profile)

	assert.Equal(t, profile.AccountID(), mapped.AccountID)
	assert.Equal(t, *profile.FirstName(), *mapped.FirstName)
	assert.Equal(t, *profile.LastName(), *mapped.LastName)
	assert.Equal(t, *profile.Phone(), *mapped.Phone)
	assert.Equal(t, *profile.AvatarURL(), *mapped.AvatarURL)
	assert.Equal(t, *profile.Bio(), *mapped.Bio)
	assert.Equal(t, profile.UpdatedAt(), mapped.UpdatedAt)
}
