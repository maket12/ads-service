//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	adapterpg "github.com/maket12/ads-service/userservice/internal/adapter/out/postgres"
	"github.com/maket12/ads-service/userservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ProfilesRepoSuite struct {
	BaseRepoSuite
	repo        *adapterpg.ProfileRepository
	testProfile *model.Profile
}

func TestProfilesRepoSuite(t *testing.T) {
	suite.Run(t, new(ProfilesRepoSuite))
}

func (s *ProfilesRepoSuite) SetupSuite() {
	s.SetupBase(1)
	s.repo = adapterpg.NewProfileRepository(s.dbClient, trmpgx.DefaultCtxGetter)
	s.ctx = context.Background()
	s.testProfile, _ = model.NewProfile(uuid.New())
}

func (s *ProfilesRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "profiles")
	s.Require().NoError(err)
}

func (s *ProfilesRepoSuite) TestCreateGet() {
	// Create at first
	err := s.repo.Create(s.ctx, s.testProfile)
	s.Require().NoError(err)

	// And then get
	profile, err := s.repo.Get(s.ctx, s.testProfile.AccountID())
	s.Require().NoError(err)
	s.Require().Exactly(s.testProfile.AccountID(), profile.AccountID())
	s.Require().WithinDuration(s.testProfile.UpdatedAt(), profile.UpdatedAt(), time.Microsecond)
}

func (s *ProfilesRepoSuite) TestCreate_Duplicate() {
	// Create a profile
	_ = s.repo.Create(s.ctx, s.testProfile)

	// Trying to create the same profile again (same account id)
	err := s.repo.Create(s.ctx, s.testProfile)
	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectAlreadyExists)
}

func (s *ProfilesRepoSuite) TestGet_NotFound() {
	// Trying to get non-existing profile
	var unexistingAccountID = uuid.New()
	_, err := s.repo.Get(s.ctx, unexistingAccountID)

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *ProfilesRepoSuite) TestUpdate() {
	// Create a profile in advance
	_ = s.repo.Create(s.ctx, s.testProfile)

	var (
		updatedProfile = *s.testProfile
		firstName      = "Vladimir"
		bio            = "digital nomad, programmer"
	)

	_ = updatedProfile.Update(
		&firstName,
		nil,
		nil,
		nil,
		&bio,
	)

	// Update
	err := s.repo.Update(s.ctx, &updatedProfile)
	s.Require().NoError(err)

	// Ensure update was successful
	profile, _ := s.repo.Get(s.ctx, updatedProfile.AccountID())
	s.Require().NotNil(profile.FirstName())
	s.Require().Equal(firstName, *profile.FirstName())
	s.Require().NotNil(profile.Bio())
	s.Require().Equal(bio, *profile.Bio())
}

func (s *ProfilesRepoSuite) TestDelete() {
	// Create a profile in advance
	_ = s.repo.Create(s.ctx, s.testProfile)

	// Then delete
	err := s.repo.Delete(s.ctx, s.testProfile.AccountID())
	s.Require().NoError(err)

	// Ensure it was deleted correctly
	_, err = s.repo.Get(s.ctx, s.testProfile.AccountID())

	s.Require().Error(err)
	s.Require().ErrorIs(err, pkgerrs.ErrObjectNotFound)
}

func (s *ProfilesRepoSuite) TestListProfiles() {
	// Create profiles
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	for _, id := range ids {
		p, _ := model.NewProfile(id)
		_ = s.repo.Create(s.ctx, p)
	}

	// Testing limit
	var (
		testLimit  = 2
		testOffset = 0
	)

	profiles, err := s.repo.ListProfiles(s.ctx, testLimit, testOffset)
	s.Require().NoError(err)
	s.Require().Len(profiles, 2, "limit should restrict result count")

	var fstFound, sndFound bool
	for i := range profiles {
		value := *profiles[i]
		if value.AccountID() == ids[0] {
			fstFound = true
		}
		if value.AccountID() == ids[1] {
			sndFound = true
		}
	}

	s.Require().Truef(fstFound, "expected account with id %v\n in %v",
		ids[0], profiles)
	s.Require().Truef(sndFound, "expected account with id %v\n in %v",
		ids[1], profiles)

	// Testing offset
	testLimit = 10
	testOffset = 2

	profiles, err = s.repo.ListProfiles(s.ctx, testLimit, testOffset)
	s.Require().NoError(err)
	s.Require().Len(profiles, 1, "offset should skip records")
	s.Require().Equal(ids[2], profiles[0].AccountID(),
		"expected account with id %v\n in %v", ids[2], profiles)
}
