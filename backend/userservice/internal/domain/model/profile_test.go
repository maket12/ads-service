package model_test

import (
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"
	"github.com/maket12/ads-service/userservice/pkg/utils"
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

		// expected* hold the values Update should persist after trimming.
		// Only checked when the corresponding input is non-nil and expect == nil.
		// If left nil while the input is non-nil, the input itself is used as
		// the expected (already-trimmed) value.
		expectedFirstName *string
		expectedLastName  *string
		expectedPhone     *string
		expectedAvatarURL *string
		expectedBio       *string
	}

	var (
		testAccID     = uuid.New()
		testFirstName = utils.VPtr(gofakeit.FirstName())
		testLastName  = utils.VPtr(gofakeit.LastName())
		testPhone     = utils.VPtr(gofakeit.Phone())
		testAvatarURL = utils.VPtr("https://img.com/a-stunning-cyberpunk-scenery.jpg")
		testBio       = utils.VPtr("programmer, seller, digital nomad")

		// Exactly at the boundary lengths accepted by Update.
		minFirstName = utils.VPtr("Ivo")  // len 3
		minLastName  = utils.VPtr("Kim")  // len 3
		minPhone     = utils.VPtr("1234") // len 4
		maxBio       = utils.VPtr(strings.Repeat("x", 512))
	)

	maxFirstNameOK := utils.VPtr(strings.Repeat("a", 15))   // len 15, valid
	tooLongFirstName := utils.VPtr(strings.Repeat("a", 16)) // len 16, invalid
	maxPhoneOK := utils.VPtr(strings.Repeat("1", 18))       // len 18, valid
	tooLongPhone := utils.VPtr(strings.Repeat("1", 19))     // len 19, invalid
	maxAvatarOK := utils.VPtr(strings.Repeat("a", 150))     // len 150, valid
	tooLongAvatar := utils.VPtr(strings.Repeat("a", 151))   // len 151, invalid
	tooLongBio := utils.VPtr(strings.Repeat("x", 513))      // len 513, invalid

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
			name:              "success - trims surrounding whitespace",
			firstName:         utils.VPtr("  " + *testFirstName + "  "),
			lastName:          utils.VPtr("\t" + *testLastName + "\t"),
			phone:             testPhone,
			avatarURL:         testAvatarURL,
			bio:               utils.VPtr("  " + *testBio + "  "),
			expect:            nil,
			expectedFirstName: testFirstName,
			expectedLastName:  testLastName,
			expectedBio:       testBio,
		},
		{
			name:      "failure - whitespace-only first name collapses below minimum",
			firstName: utils.VPtr("   "),
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: testAvatarURL,
			bio:       testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "success - first name at minimum length",
			firstName: minFirstName,
			lastName:  minLastName,
			phone:     minPhone,
			avatarURL: testAvatarURL,
			bio:       testBio,
			expect:    nil,
		},
		{
			name:      "success - first name at maximum length",
			firstName: maxFirstNameOK,
			lastName:  testLastName,
			phone:     maxPhoneOK,
			avatarURL: maxAvatarOK,
			bio:       maxBio,
			expect:    nil,
		},
		{
			name:      "failure - first name one over maximum length",
			firstName: tooLongFirstName,
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: testAvatarURL,
			bio:       testBio,
			expect:    pkgerrs.ErrValueIsInvalid,
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
			name:      "failure - phone one over maximum length",
			firstName: testFirstName,
			lastName:  testLastName,
			phone:     tooLongPhone,
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
			name:      "failure - avatar url one over maximum length",
			firstName: testFirstName,
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: tooLongAvatar,
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
		{
			name:      "failure - bio one over maximum length",
			firstName: testFirstName,
			lastName:  testLastName,
			phone:     testPhone,
			avatarURL: testAvatarURL,
			bio:       tooLongBio,
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

				var updated bool

				if tt.firstName != nil {
					want := tt.expectedFirstName
					if want == nil {
						want = tt.firstName
					}
					assert.Equal(t, want, profile.FirstName())
					updated = true
				}
				if tt.lastName != nil {
					want := tt.expectedLastName
					if want == nil {
						want = tt.lastName
					}
					assert.Equal(t, want, profile.LastName())
					updated = true
				}
				if tt.phone != nil {
					want := tt.expectedPhone
					if want == nil {
						want = tt.phone
					}
					assert.Equal(t, want, profile.Phone())
					updated = true
				}
				if tt.avatarURL != nil {
					want := tt.expectedAvatarURL
					if want == nil {
						want = tt.avatarURL
					}
					assert.Equal(t, want, profile.AvatarURL())
					updated = true
				}
				if tt.bio != nil {
					want := tt.expectedBio
					if want == nil {
						want = tt.bio
					}
					assert.Equal(t, want, profile.Bio())
					updated = true
				}

				if updated {
					assert.NotEqual(t, updAt, profile.UpdatedAt())
				}
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Equal(t, updAt, profile.UpdatedAt())
			}
		})
	}
}
