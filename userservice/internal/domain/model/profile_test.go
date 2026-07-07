package model_test

import (
	"testing"
	"time"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"github.com/maket12/ads-service/pkg/utils"
	"github.com/maket12/ads-service/userservice/internal/domain/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfile(t *testing.T) {
	type testCase struct {
		name      string
		accountID uuid.UUID
		expect    error
	}

	var tests = []testCase{
		{
			name:      "success",
			accountID: uuid.New(),
			expect:    nil,
		},
		{
			name:      "failure - invalid account id",
			accountID: uuid.Nil,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, err := model.NewProfile(tt.accountID)
			if tt.expect == nil {
				require.NoError(t, err)
				assert.Equal(t, tt.accountID, profile.AccountID())
				assert.NotNil(t, profile.UpdatedAt())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, profile)
			}
		})
	}
}

func TestProfile_Update(t *testing.T) {
	type testCase struct {
		name      string
		firstName *string
		lastName  *string
		phone     *string
		avatarURL *string
		bio       *string
		expect    error
	}

	var (
		testAccID     = uuid.New()
		testFirstName = utils.VPtr("Vladimir")
		testLastName  = utils.VPtr("Ziabkin")
		testPhone     = utils.VPtr("+79137918725")
		testAvatarURL = utils.VPtr("https://img.com/a-stunning-cyberpunk-scenery.jpg")
		testBio       = utils.VPtr("programmer, seller, digital nomad")
	)

	var tests = []testCase{
		{
			name:      "success",
			firstName: testFirstName,
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: testAvatarURL,
			bio:       testBio,
			expect:    nil,
		},
		{
			name:      "success - nothing to update",
			firstName: nil,
			lastName:  nil,
			phone:     nil,
			avatarURL: nil,
			bio:       nil,
			expect:    nil,
		},
		{
			name:      "failure - invalid first name",
			firstName: utils.VPtr("A"),
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: testAvatarURL,
			bio:       testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid last name",
			firstName: testFirstName,
			lastName:  utils.VPtr("A"),
			phone:     testPhone,
			avatarURL: testAvatarURL,
			bio:       testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid phone",
			firstName: testFirstName,
			lastName:  testLastName,
			phone:     utils.VPtr(""),
			avatarURL: testAvatarURL,
			bio:       testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid avatar url",
			firstName: testFirstName,
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: utils.VPtr(""),
			bio:       testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid bio",
			firstName: testFirstName,
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: testAvatarURL,
			bio:       utils.VPtr(""),
			expect:    pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile, _ := model.NewProfile(testAccID)
			updAt := profile.UpdatedAt()

			// Wait to update current time
			time.Sleep(time.Millisecond)

			err := profile.Update(
				tt.firstName, tt.lastName,
				tt.phone, tt.avatarURL, tt.bio,
			)
			if tt.expect == nil {
				require.NoError(t, err)

				if tt.firstName != nil {
					assert.Equal(t, tt.firstName, profile.FirstName())
				}
				if tt.lastName != nil {
					assert.Equal(t, tt.lastName, profile.LastName())
				}
				if tt.phone != nil {
					assert.Equal(t, tt.phone, profile.Phone())
				}
				if tt.avatarURL != nil {
					assert.Equal(t, tt.avatarURL, profile.AvatarURL())
				}
				if tt.bio != nil {
					assert.Equal(t, tt.bio, profile.Bio())
				}

				assert.NotEqual(t, updAt, profile.UpdatedAt())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Equal(t, updAt, profile.UpdatedAt())
			}
		})
	}
}
