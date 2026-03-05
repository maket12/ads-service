package mongodb

import (
	"context"
	"errors"

	pkgmongo "github.com/maket12/ads-service/pkg/mongodb"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MediaDocument struct {
	AdID   string   `bson:"ad_id"`
	Images []string `bson:"images"`
}

type MediaRepositoryConfig struct {
	mongoClient    *pkgmongo.Client
	collectionName string
}

func NewMediaRepositoryConfig(
	mongoClient *pkgmongo.Client,
	collectionName string,
) *MediaRepositoryConfig {
	return &MediaRepositoryConfig{
		mongoClient:    mongoClient,
		collectionName: collectionName,
	}
}

func (c *MediaRepositoryConfig) collection() *mongo.Collection {
	return c.mongoClient.Database.Collection(c.collectionName)
}

type MediaRepository struct {
	collection *mongo.Collection
}

func NewMediaRepository(mediaRepoCfg *MediaRepositoryConfig) *MediaRepository {
	return &MediaRepository{collection: mediaRepoCfg.collection()}
}

// Save method will update if record already exists and add otherwise
func (r *MediaRepository) Save(ctx context.Context, adID uuid.UUID, images []string) error {
	filter := bson.M{"ad_id": adID.String()}
	update := bson.M{
		"$set": MediaDocument{
			AdID:   adID.String(),
			Images: images,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// Get method returns images if they are in database, otherwise an empty list
func (r *MediaRepository) Get(ctx context.Context, adID uuid.UUID) ([]string, error) {
	var doc MediaDocument
	err := r.collection.FindOne(ctx, bson.M{"ad_id": adID.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []string{}, nil
		}
		return nil, err
	}
	return doc.Images, nil
}

// Delete method delete images related with given ad_id
func (r *MediaRepository) Delete(ctx context.Context, adID uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"ad_id": adID.String()})
	if err != nil {
		return err
	}
	return nil
}
