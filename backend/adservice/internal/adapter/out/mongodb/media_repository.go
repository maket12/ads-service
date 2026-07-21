package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/maket12/ads-service/adservice/internal/domain/port"
	pkgmongo "github.com/maket12/ads-service/adservice/pkg/mongodb"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MediaDocument struct {
	AdID      string      `bson:"ad_id"`
	Images    []ImageMeta `bson:"images"`
	CreatedAt time.Time   `bson:"created_at"`
	UpdatedAt time.Time   `bson:"updated_at"`
}

type ImageMeta struct {
	ID         string    `bson:"id"`
	URL        string    `bson:"url"`
	Width      int       `bson:"width,omitempty"`
	Height     int       `bson:"height,omitempty"`
	SizeBytes  int64     `bson:"size_bytes,omitempty"`
	Format     string    `bson:"format,omitempty"` // "jpeg", "png", "webp", etc.
	UploadedAt time.Time `bson:"uploaded_at"`
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
func (r *MediaRepository) Save(ctx context.Context, adID uuid.UUID, images []port.ImageInput) error {
	now := time.Now()

	imgMetas := make([]ImageMeta, len(images))
	for i, img := range images {
		imgMetas[i] = ImageMeta{
			ID:         img.ID,
			URL:        img.URL,
			Width:      img.Width,
			Height:     img.Height,
			SizeBytes:  img.SizeBytes,
			Format:     img.Format,
			UploadedAt: now,
		}
	}

	filter := bson.M{"ad_id": adID.String()}
	update := bson.M{
		"$set": bson.M{
			"images":     imgMetas,
			"updated_at": now,
		},
		"$setOnInsert": bson.M{
			"ad_id":      adID.String(),
			"created_at": now,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// Get method returns images if they are in database, otherwise an empty list
func (r *MediaRepository) Get(ctx context.Context, adID uuid.UUID) ([]port.ImageRef, error) {
	var doc MediaDocument
	err := r.collection.FindOne(ctx, bson.M{"ad_id": adID.String()}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []port.ImageRef{}, nil
		}
		return nil, err
	}

	refs := make([]port.ImageRef, len(doc.Images))
	for i, img := range doc.Images {
		refs[i] = port.ImageRef{
			ID:     img.ID,
			URL:    img.URL,
			Width:  img.Width,
			Height: img.Height,
		}
	}
	return refs, nil
}

// Delete method delete images related with given ad_id
func (r *MediaRepository) Delete(ctx context.Context, adID uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"ad_id": adID.String()})
	if err != nil {
		return err
	}
	return nil
}
