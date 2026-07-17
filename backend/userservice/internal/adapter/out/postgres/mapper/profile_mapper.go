package mapper

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/postgres/sqlc"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
)

func MapProfileToSQLCCreate(profile *model.Profile) sqlc.CreateProfileParams {
	var (
		firstName pgtype.Text
		lastName  pgtype.Text
		phone     pgtype.Text
		avatarURL pgtype.Text
		bio       pgtype.Text
	)
	if profile.FirstName() != nil {
		firstName = pgtype.Text{
			String: *profile.FirstName(),
			Valid:  true,
		}
	}
	if profile.LastName() != nil {
		lastName = pgtype.Text{
			String: *profile.LastName(),
			Valid:  true,
		}
	}
	if profile.Phone() != nil {
		phone = pgtype.Text{
			String: *profile.Phone(),
			Valid:  true,
		}
	}
	if profile.AvatarURL() != nil {
		avatarURL = pgtype.Text{
			String: *profile.AvatarURL(),
			Valid:  true,
		}
	}
	if profile.Bio() != nil {
		bio = pgtype.Text{
			String: *profile.Bio(),
			Valid:  true,
		}
	}

	return sqlc.CreateProfileParams{
		AccountID: pgtype.UUID{
			Bytes: profile.AccountID(),
			Valid: true,
		},
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		AvatarUrl: avatarURL,
		Bio:       bio,
		UpdatedAt: pgtype.Timestamptz{
			Time:  profile.UpdatedAt(),
			Valid: true,
		},
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
		rawProfile.AccountID.Bytes,
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
		firstName pgtype.Text
		lastName  pgtype.Text
		phone     pgtype.Text
		avatarURL pgtype.Text
		bio       pgtype.Text
	)
	if profile.FirstName() != nil {
		firstName = pgtype.Text{
			String: *profile.FirstName(),
			Valid:  true,
		}
	}
	if profile.LastName() != nil {
		lastName = pgtype.Text{
			String: *profile.LastName(),
			Valid:  true,
		}
	}
	if profile.Phone() != nil {
		phone = pgtype.Text{
			String: *profile.Phone(),
			Valid:  true,
		}
	}
	if profile.AvatarURL() != nil {
		avatarURL = pgtype.Text{
			String: *profile.AvatarURL(),
			Valid:  true,
		}
	}
	if profile.Bio() != nil {
		bio = pgtype.Text{
			String: *profile.Bio(),
			Valid:  true,
		}
	}

	return sqlc.UpdateProfileParams{
		AccountID: pgtype.UUID{
			Bytes: profile.AccountID(),
			Valid: true,
		},
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		AvatarUrl: avatarURL,
		Bio:       bio,
		UpdatedAt: pgtype.Timestamptz{
			Time:  profile.UpdatedAt(),
			Valid: true,
		},
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
