package mongodb_test

import (
	"context"
	adaptermongodb "github.com/maket12/ads-service/adservice/internal/adapter/out/mongodb"
	pkgmongodb "github.com/maket12/ads-service/pkg/mongodb"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type MediaRepoSuite struct {
	suite.Suite
	dbClient *pkgmongodb.Client
	repo     *adaptermongodb.MediaRepository
	ctx      context.Context
}

func TestMediaRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(MediaRepoSuite))
}

func (s *MediaRepoSuite) SetupSuite() {
	s.ctx = context.Background()

	cfg := pkgmongodb.NewConfig(
		"localhost", 27017,
		"test", "test",
		"test-mongo",
	)

	dbClient, err := pkgmongodb.NewClient(s.ctx, cfg)
	s.Require().NoError(err)

	s.dbClient = dbClient

	repoCfg := adaptermongodb.NewMediaRepositoryConfig(
		s.dbClient, "test-images",
	)
	s.repo = adaptermongodb.NewMediaRepository(repoCfg)
}

func (s *MediaRepoSuite) SetupTest() {
	err := s.dbClient.Database.Collection("test-images").Drop(s.ctx)
	s.Require().NoError(err)
}

func (s *MediaRepoSuite) TearDownSuite() {
	err := s.dbClient.Close(s.ctx)
	s.Require().NoError(err)
}

func (s *MediaRepoSuite) TestSaveGet() {
	// Prepare test data
	var (
		testAdID   = uuid.New()
		testImages = []string{
			"https://storage.com/1.jpg",
			"https://storage.com/2.jpg",
		}
	)

	// Save
	err := s.repo.Save(s.ctx, testAdID, testImages)
	s.Require().NoError(err)

	// And then get
	images, err := s.repo.Get(s.ctx, testAdID)
	s.Require().NoError(err)
	s.Require().NotNil(images)
	s.Require().ElementsMatch(testImages, images)
}

func (s *MediaRepoSuite) TestGet_NotFound() {
	// Trying to get non-existing data
	var unexistingAdID = uuid.New()

	images, err := s.repo.Get(s.ctx, unexistingAdID)
	s.Require().NoError(err)
	s.Require().Empty(images)
}

func (s *MediaRepoSuite) TestDelete() {
	// Prepare test data
	var (
		testAdID   = uuid.New()
		testImages = []string{
			"https://storage.com/1.jpg",
			"https://storage.com/2.jpg",
		}
	)

	// Save
	_ = s.repo.Save(s.ctx, testAdID, testImages)

	// Then delete
	err := s.repo.Delete(s.ctx, testAdID)
	s.Require().NoError(err)

	// Check deletion was correct
	images, _ := s.repo.Get(s.ctx, testAdID)
	s.Require().Empty(images)
}
