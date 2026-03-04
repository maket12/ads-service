package mapper

import (
	"database/sql"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"time"
)

func MapProfileToSQLCCreate(profile *model.Profile) sqlc.CreateProfileParams {
	var (
		firstName sql.NullString
		lastName  sql.NullString
		phone     sql.NullString
		avatarURL sql.NullString
		bio       sql.NullString
	)
	if profile.FirstName() != nil {
		firstName = sql.NullString{
			String: *profile.FirstName(),
			Valid:  true,
		}
	}
	if profile.LastName() != nil {
		lastName = sql.NullString{
			String: *profile.LastName(),
			Valid:  true,
		}
	}
	if profile.Phone() != nil {
		phone = sql.NullString{
			String: *profile.Phone(),
			Valid:  true,
		}
	}
	if profile.AvatarURL() != nil {
		avatarURL = sql.NullString{
			String: *profile.AvatarURL(),
			Valid:  true,
		}
	}
	if profile.Bio() != nil {
		bio = sql.NullString{
			String: *profile.Bio(),
			Valid:  true,
		}
	}

	updatedAt := sql.NullTime{
		Time:  profile.UpdatedAt(),
		Valid: true,
	}

	return sqlc.CreateProfileParams{
		AccountID: profile.AccountID(),
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		AvatarUrl: avatarURL,
		Bio:       bio,
		UpdatedAt: updatedAt,
	}
}

func MapSQLCToProfile(rawProfile sqlc.Profile) *model.Profile {
	var (
		firstName *string
		lastName  *string
		phone     *string
		avatarURL *string
		bio       *string
		updatedAt time.Time
	)
	if rawProfile.FirstName.Valid {
		firstName = &rawProfile.FirstName.String
	}
	if rawProfile.LastName.Valid {
		lastName = &rawProfile.LastName.String
	}
	if rawProfile.Phone.Valid {
		phone = &rawProfile.Phone.String
	}
	if rawProfile.AvatarUrl.Valid {
		avatarURL = &rawProfile.AvatarUrl.String
	}
	if rawProfile.Bio.Valid {
		bio = &rawProfile.Bio.String
	}
	if rawProfile.UpdatedAt.Valid {
		updatedAt = rawProfile.UpdatedAt.Time
	}

	return model.RestoreProfile(
		rawProfile.AccountID,
		firstName,
		lastName,
		phone,
		avatarURL,
		bio,
		updatedAt,
	)
}

func MapProfileToSQLCUpdate(profile *model.Profile) sqlc.UpdateProfileParams {
	var (
		firstName sql.NullString
		lastName  sql.NullString
		phone     sql.NullString
		avatarURL sql.NullString
		bio       sql.NullString
	)
	if profile.FirstName() != nil {
		firstName = sql.NullString{
			String: *profile.FirstName(),
			Valid:  true,
		}
	}
	if profile.LastName() != nil {
		lastName = sql.NullString{
			String: *profile.LastName(),
			Valid:  true,
		}
	}
	if profile.Phone() != nil {
		phone = sql.NullString{
			String: *profile.Phone(),
			Valid:  true,
		}
	}
	if profile.AvatarURL() != nil {
		avatarURL = sql.NullString{
			String: *profile.AvatarURL(),
			Valid:  true,
		}
	}
	if profile.Bio() != nil {
		bio = sql.NullString{
			String: *profile.Bio(),
			Valid:  true,
		}
	}

	updatedAt := sql.NullTime{
		Time:  profile.UpdatedAt(),
		Valid: true,
	}

	return sqlc.UpdateProfileParams{
		AccountID: profile.AccountID(),
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		AvatarUrl: avatarURL,
		Bio:       bio,
		UpdatedAt: updatedAt,
	}
}

func MapToSQLCList(limit, offset int) sqlc.ListProfilesParams {
	return sqlc.ListProfilesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}
}

func MapSQLCToProfilesList(rawProfiles []sqlc.Profile) []*model.Profile {
	profiles := make([]*model.Profile, 0, len(rawProfiles))
	for _, rawProfile := range rawProfiles {
		mappedProfile := MapSQLCToProfile(rawProfile)
		profiles = append(profiles, mappedProfile)
	}
	return profiles
}
