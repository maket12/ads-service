package model

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"
	"github.com/maket12/ads-service/userservice/pkg/utils"
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
	if firstName != nil {
		fNameLen := utf8.RuneCountInString(strings.TrimSpace(*firstName))
		if fNameLen < 3 || fNameLen > 15 {
			return pkgerrs.NewValueInvalidError("first_name")
		}
	}

	if lastName != nil {
		lNameLen := utf8.RuneCountInString(strings.TrimSpace(*lastName))
		if lNameLen < 3 || lNameLen > 15 {
			return pkgerrs.NewValueInvalidError("last_name")
		}
	}

	if phone != nil {
		phoneLen := utf8.RuneCountInString(strings.TrimSpace(*phone))
		if phoneLen < 4 || phoneLen > 18 {
			return pkgerrs.NewValueInvalidError("phone")
		}
	}

	if avatarURL != nil {
		avatarLen := utf8.RuneCountInString(strings.TrimSpace(*avatarURL))
		if avatarLen == 0 || avatarLen > 150 {
			return pkgerrs.NewValueInvalidError("avatar_url")
		}
	}

	if bio != nil {
		bioLen := utf8.RuneCountInString(strings.TrimSpace(*bio))
		if bioLen == 0 || bioLen > 512 {
			return pkgerrs.NewValueInvalidError("bio")
		}
	}

	var changed bool

	if firstName != nil {
		p.firstName = utils.VPtr(strings.TrimSpace(*firstName))
		changed = true
	}
	if lastName != nil {
		p.lastName = utils.VPtr(strings.TrimSpace(*lastName))
		changed = true
	}
	if phone != nil {
		p.phone = utils.VPtr(strings.TrimSpace(*phone))
		changed = true
	}
	if avatarURL != nil {
		p.avatarURL = utils.VPtr(strings.TrimSpace(*avatarURL))
		changed = true
	}
	if bio != nil {
		p.bio = utils.VPtr(strings.TrimSpace(*bio))
		changed = true
	}

	if changed {
		p.updatedAt = time.Now()
	}

	return nil
}
