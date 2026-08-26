//go:build integration

package mongodb_test

import (
	"context"
	"log"
	"os"
	"testing"

	adaptermongodb "github.com/maket12/ads-service/backend/adservice/v2/internal/adapter/out/mongodb"
	pkgmongodb "github.com/maket12/ads-service/backend/adservice/v2/pkg/mongodb"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

var (
	globalContainer *pkgmongodb.TestContainer
	globalClient    *pkgmongodb.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	mongoContainer, err := pkgmongodb.StartTestContainer(ctx)
	if err != nil {
		log.Fatalf("Could not start mongodb: %v", err)
	}
	globalContainer = mongoContainer

	client, err := pkgmongodb.NewClient(ctx, mongoContainer.Config)
	if err != nil {
		log.Fatalf("Could not connect to mongodb: %v", err)
	}
	globalClient = client

	code := m.Run()

	_ = globalClient.Close(ctx)
	_ = globalContainer.Close(ctx)

	os.Exit(code)
}

type MediaRepoSuite struct {
	suite.Suite
	mongoContainer *pkgmongodb.TestContainer
	dbClient       *pkgmongodb.Client
	repo           *adaptermongodb.MediaRepository
	ctx            context.Context
}

func TestMediaRepoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(MediaRepoSuite))
}

func (s *MediaRepoSuite) SetupSuite() {
	s.ctx = context.Background()
	s.mongoContainer = globalContainer
	s.dbClient = globalClient

	repoCfg := adaptermongodb.NewMediaRepositoryConfig(
		s.dbClient, "test-images",
	)
	s.repo = adaptermongodb.NewMediaRepository(repoCfg)
}

func (s *MediaRepoSuite) SetupTest() {
	err := s.mongoContainer.DropCollections(s.ctx, "test-images")
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
