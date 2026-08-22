package grpc_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/adservice/internal/adapter/in/grpc"
	"github.com/maket12/ads-service/adservice/internal/app/dto"
	"github.com/maket12/ads-service/adservice/pkg/generated/ad_v1"
	"github.com/stretchr/testify/assert"
)

func TestMapCreateAdPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	desc := "New description"
	req := &ad_v1.CreateAdRequest{
		Title:       "Test Title",
		Description: &desc,
		Price:       1500,
		Images:      []string{"img1.jpg"},
	}

	result := grpc.MapCreateAdPbToDTO(req, sellerID)

	assert.Equal(t, sellerID, result.SellerID)
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, req.Description, result.Description)
	assert.Equal(t, req.Price, result.Price)
	assert.Equal(t, req.Images, result.Images)
}

func TestMapCreateAdDTOToPb(t *testing.T) {
	adID := uuid.New()
	out := dto.CreateAdOutput{AdID: adID}

	result := grpc.MapCreateAdDTOToPb(out)

	assert.Equal(t, adID.String(), result.AdId)
}

func TestMapGetAdPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	adID := uuid.New()
	req := &ad_v1.GetAdRequest{AdId: adID.String()}

	result := grpc.MapGetAdPbToDTO(req, sellerID)

	assert.Equal(t, adID, result.AdID)
	assert.Equal(t, sellerID, result.SellerID)
}

func TestMapGetAdDTOToPb(t *testing.T) {
	adID := uuid.New()
	now := time.Now().Truncate(time.Second)
	out := dto.GetAdOutput{
		AdID:      adID,
		Title:     "Wrapped Ad",
		CreatedAt: now,
	}

	result := grpc.MapGetAdDTOToPb(out)

	assert.Equal(t, adID.String(), result.Ad.AdId)
	assert.Equal(t, "Wrapped Ad", result.Ad.Title)
}

func TestMapUpdateAdPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	adID := uuid.New()
	title := "Updated Title"
	desc := "Updated Desc"
	var price int64 = 9999

	req := &ad_v1.UpdateAdRequest{
		AdId:        adID.String(),
		Title:       &title,
		Description: &desc,
		Price:       &price,
		Images:      []string{"updated.jpg"},
	}

	result := grpc.MapUpdateAdPbToDTO(req, sellerID)

	assert.Equal(t, adID, result.AdID)
	assert.Equal(t, sellerID, result.SellerID)
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, req.Description, result.Description)
	assert.Equal(t, req.Price, result.Price)
	assert.Equal(t, req.Images, result.Images)
}

func TestMapUpdateAdDTOToPb(t *testing.T) {
	out := dto.UpdateAdOutput{Success: true}

	result := grpc.MapUpdateAdDTOToPb(out)

	assert.True(t, result.Success)
}

func TestMapPublishAdPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	adID := uuid.New()
	req := &ad_v1.PublishAdRequest{AdId: adID.String()}

	result := grpc.MapPublishAdPbToDTO(req, sellerID)

	assert.Equal(t, adID, result.AdID)
	assert.Equal(t, sellerID, result.SellerID)
}

func TestMapPublishAdDTOToPb(t *testing.T) {
	out := dto.PublishAdOutput{Success: true}

	result := grpc.MapPublishAdDTOToPb(out)

	assert.True(t, result.Success)
}

func TestMapRejectAdPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	adID := uuid.New()
	req := &ad_v1.RejectAdRequest{AdId: adID.String()}

	result := grpc.MapRejectAdPbToDTO(req, sellerID)

	assert.Equal(t, adID, result.AdID)
	assert.Equal(t, sellerID, result.SellerID)
}

func TestMapRejectAdDTOToPb(t *testing.T) {
	out := dto.RejectAdOutput{Success: true}

	result := grpc.MapRejectAdDTOToPb(out)

	assert.True(t, result.Success)
}

func TestMapDeleteAdPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	adID := uuid.New()
	req := &ad_v1.DeleteAdRequest{AdId: adID.String()}

	result := grpc.MapDeleteAdPbToDTO(req, sellerID)

	assert.Equal(t, adID, result.AdID)
	assert.Equal(t, sellerID, result.SellerID)
}

func TestMapDeleteAdDTOToPb(t *testing.T) {
	out := dto.DeleteAdOutput{Success: true}

	result := grpc.MapDeleteAdDTOToPb(out)

	assert.True(t, result.Success)
}

func TestMapDeleteAllAdsPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	req := &ad_v1.DeleteAllAdsRequest{SellerId: sellerID.String()}

	result := grpc.MapDeleteAllAdsPbToDTO(req)

	assert.Equal(t, sellerID, result.SellerID)
}

func TestMapDeleteAllAdsDTOToPb(t *testing.T) {
	out := dto.DeleteAllAdsOutput{Success: true}

	result := grpc.MapDeleteAllAdsDTOToPb(out)

	assert.True(t, result.Success)
}

func TestMapListAdsPbToDTO(t *testing.T) {
	sellerID := uuid.New()
	req := &ad_v1.ListAdsRequest{}

	result := grpc.MapListAdsPbToDTO(req, sellerID)

	assert.Equal(t, sellerID, result.SellerID)
}

func TestMapListAdsDTOToPb(t *testing.T) {
	adID := uuid.New()
	now := time.Now().Truncate(time.Second)

	ad := dto.Ad{
		AdID:      adID,
		Title:     "Listed Ad",
		CreatedAt: now,
	}
	out := dto.ListSellerAdsOutput{Ads: []dto.Ad{ad}}

	result := grpc.MapListAdsDTOToPb(out)

	assert.Len(t, result.Ads, 1)
	assert.Equal(t, adID.String(), result.Ads[0].AdId)
	assert.Equal(t, "Listed Ad", result.Ads[0].Title)
}
