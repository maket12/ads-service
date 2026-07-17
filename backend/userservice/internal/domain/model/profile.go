package model

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"
)

// ================ Rich model for Profile ================

type Profile struct {
	accountID uuid.UUID
	firstName *string
	lastName  *string
	phone     *string
	avatarURL *string
	bio       *string
	updatedAt time.Time
}

func NewProfile(accountID uuid.UUID) (*Profile, error) {
	if accountID == uuid.Nil {
		return nil, pkgerrs.NewValueInvalidError("account_id")
	}
	return &Profile{
		accountID: accountID,
		updatedAt: time.Now(),
	}, nil
}

func RestoreProfile(accountID uuid.UUID,
	firstName, lastName, phone, avatarURL, bio *string,
	updatedAt time.Time,
) *Profile {
	return &Profile{
		accountID: accountID,
		firstName: firstName,
		lastName:  lastName,
		phone:     phone,
		avatarURL: avatarURL,
		bio:       bio,
		updatedAt: updatedAt,
	}
}

// ================ Read-Only ================

func (p *Profile) AccountID() uuid.UUID { return p.accountID }
func (p *Profile) FirstName() *string   { return p.firstName }
func (p *Profile) LastName() *string    { return p.lastName }
func (p *Profile) Phone() *string       { return p.phone }
func (p *Profile) AvatarURL() *string   { return p.avatarURL }
func (p *Profile) Bio() *string         { return p.bio }
func (p *Profile) UpdatedAt() time.Time { return p.updatedAt }

// ================ Mutation ================

func (p *Profile) Update(firstName, lastName, phone, avatarURL, bio *string) error {
	if firstName != nil &&
		utf8.RuneCountInString(strings.TrimSpace(*firstName)) < 3 {
		return pkgerrs.NewValueInvalidError("first_name")
	}
	if lastName != nil &&
		utf8.RuneCountInString(strings.TrimSpace(*lastName)) < 3 {
		return pkgerrs.NewValueInvalidError("last_name")
	}
	if phone != nil &&
		utf8.RuneCountInString(strings.TrimSpace(*phone)) < 4 {
		return pkgerrs.NewValueInvalidError("phone")
	}
	if avatarURL != nil && *avatarURL == "" {
		return pkgerrs.NewValueInvalidError("avatar_url")
	}
	if bio != nil && *bio == "" {
		return pkgerrs.NewValueInvalidError("bio")
	}

	var changed bool

	if firstName != nil {
		p.firstName = firstName
		changed = true
	}
	if lastName != nil {
		p.lastName = lastName
		changed = true
	}
	if phone != nil {
		p.phone = phone
		changed = true
	}
	if avatarURL != nil {
		p.avatarURL = avatarURL
		changed = true
	}
	if bio != nil {
		p.bio = bio
		changed = true
	}

	if changed {
		p.updatedAt = time.Now()
	}

	return nil
}
