package grpc

import (
	dto2 "github.com/maket12/ads-service/backend/adservice/internal/app/dto"
	"github.com/maket12/ads-service/pkg/generated/ad_v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapCreateAdPbToDTO(req *ad_v1.CreateAdRequest, sellerID uuid.UUID) dto2.CreateAdInput {
	return dto2.CreateAdInput{
		SellerID:    sellerID,
		Title:       req.GetTitle(),
		Description: req.Description,
		Price:       req.GetPrice(),
		Images:      req.GetImages(),
	}
}

func MapCreateAdDTOToPb(out dto2.CreateAdOutput) *ad_v1.CreateAdResponse {
	return &ad_v1.CreateAdResponse{AdId: out.AdID.String()}
}

func MapGetAdPbToDTO(req *ad_v1.GetAdRequest, sellerID uuid.UUID) dto2.GetAdInput {
	adID, _ := uuid.Parse(req.GetAdId())
	return dto2.GetAdInput{
		AdID:     adID,
		SellerID: sellerID,
	}
}

func MapGetAdDTOToPb(out dto2.GetAdOutput) *ad_v1.GetAdResponse {
	return &ad_v1.GetAdResponse{
		AdId:        out.AdID.String(),
		SellerId:    out.SellerID.String(),
		Title:       out.Title,
		Description: out.Description,
		Price:       out.Price,
		Status:      out.Status,
		Images:      out.Images,
		CreatedAt:   timestamppb.New(out.CreatedAt),
		UpdatedAt:   timestamppb.New(out.UpdatedAt),
	}
}

func MapUpdateAdPbToDTO(req *ad_v1.UpdateAdRequest, sellerID uuid.UUID) dto2.UpdateAdInput {
	adID, _ := uuid.Parse(req.GetAdId())
	return dto2.UpdateAdInput{
		AdID:        adID,
		SellerID:    sellerID,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Images:      req.Images,
	}
}

func MapUpdateAdDTOToPb(out dto2.UpdateAdOutput) *ad_v1.UpdateAdResponse {
	return &ad_v1.UpdateAdResponse{Success: out.Success}
}

func MapPublishAdPbToDTO(req *ad_v1.PublishAdRequest, sellerID uuid.UUID) dto2.PublishAdInput {
	adID, _ := uuid.Parse(req.GetAdId())
	return dto2.PublishAdInput{
		AdID:     adID,
		SellerID: sellerID,
	}
}

func MapPublishAdDTOToPb(out dto2.PublishAdOutput) *ad_v1.PublishAdResponse {
	return &ad_v1.PublishAdResponse{Success: out.Success}
}

func MapRejectAdPbToDTO(req *ad_v1.RejectAdRequest, sellerID uuid.UUID) dto2.RejectAdInput {
	adID, _ := uuid.Parse(req.GetAdId())
	return dto2.RejectAdInput{
		AdID:     adID,
		SellerID: sellerID,
	}
}

func MapRejectAdDTOToPb(out dto2.RejectAdOutput) *ad_v1.RejectAdResponse {
	return &ad_v1.RejectAdResponse{Success: out.Success}
}

func MapDeleteAdPbToDTO(req *ad_v1.DeleteAdRequest, sellerID uuid.UUID) dto2.DeleteAdInput {
	adID, _ := uuid.Parse(req.GetAdId())
	return dto2.DeleteAdInput{
		AdID:     adID,
		SellerID: sellerID,
	}
}

func MapDeleteAdDTOToPb(out dto2.DeleteAdOutput) *ad_v1.DeleteAdResponse {
	return &ad_v1.DeleteAdResponse{Success: out.Success}
}

func MapDeleteAllAdsPbToDTO(req *ad_v1.DeleteAllAdsRequest) dto2.DeleteAllAdsInput {
	sellerID, _ := uuid.Parse(req.GetSellerId())
	return dto2.DeleteAllAdsInput{SellerID: sellerID}
}

func MapDeleteAllAdsDTOToPb(out dto2.DeleteAllAdsOutput) *ad_v1.DeleteAllAdsResponse {
	return &ad_v1.DeleteAllAdsResponse{Success: out.Success}
}
