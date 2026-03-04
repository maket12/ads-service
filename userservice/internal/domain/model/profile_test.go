package model_test

import (
	pkgerrs "github.com/maket12/ads-service/pkg/errs"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProfile(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	// Helper
	strPtr := func(s string) *string { return &s }

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
		testFirstName = "Vladimir"
		testLastName  = "Ziabkin"
		testPhone     = "+79137918725"
		testAvatarURL = "https://img.com/a-stunning-cyberpunk-scenery.jpg"
		testBio       = "programmer, seller, digital nomad"
	)

	var tests = []testCase{
		{
			name:      "success",
			firstName: &testFirstName,
			lastName:  &testLastName,
			phone:     &testPhone,
			avatarURL: &testAvatarURL,
			bio:       &testBio,
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
			firstName: strPtr("A"),
			lastName:  &testLastName,
			phone:     &testPhone,
			avatarURL: &testAvatarURL,
			bio:       &testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid last name",
			firstName: &testFirstName,
			lastName:  strPtr("A"),
			phone:     &testPhone,
			avatarURL: &testAvatarURL,
			bio:       &testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid phone",
			firstName: &testFirstName,
			lastName:  &testLastName,
			phone:     strPtr(""),
			avatarURL: &testAvatarURL,
			bio:       &testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid avatar url",
			firstName: &testFirstName,
			lastName:  &testLastName,
			phone:     &testPhone,
			avatarURL: strPtr(""),
			bio:       &testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "failure - invalid bio",
			firstName: &testFirstName,
			lastName:  &testLastName,
			phone:     &testPhone,
			avatarURL: &testAvatarURL,
			bio:       strPtr(""),
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
