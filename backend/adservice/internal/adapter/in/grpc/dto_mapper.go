package grpc

import (
	"github.com/maket12/ads-service/backend/adservice/api/proto/generated/ad_v1"
	"github.com/maket12/ads-service/backend/adservice/internal/app/dto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapCreateAdPbToDTO(req *ad_v1.CreateAdRequest, sellerID uuid.UUID) dto.CreateAdInput {
	return dto.CreateAdInput{
		SellerID:    sellerID,
		Title:       req.GetTitle(),
		Description: req.Description,
		Price:       req.GetPrice(),
		Category:    req.Category,
		Images:      req.GetImages(),
	}
}

func MapCreateAdDTOToPb(out dto.CreateAdOutput) *ad_v1.CreateAdResponse {
	return &ad_v1.CreateAdResponse{Id: out.AdID.String()}
}

func MapGetAdPbToDTO(req *ad_v1.GetAdRequest, sellerID uuid.UUID) dto.GetAdInput {
	adID, _ := uuid.Parse(req.GetId())
	return dto.GetAdInput{
		AdID:     adID,
		SellerID: sellerID,
	}
}

func mapAdDTOToPB(ad dto.Ad) *ad_v1.Ad {
	var updatedAtProto *timestamppb.Timestamp
	if ad.UpdatedAt != nil {
		updatedAtProto = timestamppb.New(*ad.UpdatedAt)
	}

	return &ad_v1.Ad{
		Id:          ad.AdID.String(),
		SellerId:    ad.SellerID.String(),
		Title:       ad.Title,
		Description: ad.Description,
		Price:       ad.Price,
		Category:    ad.Category,
		Status:      ad.Status,
		Images:      ad.Images,
		CreatedAt:   timestamppb.New(ad.CreatedAt),
		UpdatedAt:   updatedAtProto,
	}
}

func MapGetAdDTOToPb(out dto.GetAdOutput) *ad_v1.GetAdResponse {
	return &ad_v1.GetAdResponse{Ad: mapAdDTOToPB(dto.Ad(out))}
}

func MapUpdateAdPbToDTO(req *ad_v1.UpdateAdRequest, sellerID uuid.UUID) dto.UpdateAdInput {
	adID, _ := uuid.Parse(req.GetId())
	return dto.UpdateAdInput{
		AdID:        adID,
		SellerID:    sellerID,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		Images:      req.Images,
	}
}

func MapUpdateAdDTOToPb(out dto.UpdateAdOutput) *ad_v1.UpdateAdResponse {
	return &ad_v1.UpdateAdResponse{Success: out.Success}
}

func MapPublishAdPbToDTO(req *ad_v1.PublishAdRequest) dto.PublishAdInput {
	adID, _ := uuid.Parse(req.GetId())
	return dto.PublishAdInput{
		AdID: adID,
	}
}

func MapPublishAdDTOToPb(out dto.PublishAdOutput) *ad_v1.PublishAdResponse {
	return &ad_v1.PublishAdResponse{Success: out.Success}
}

func MapRejectAdPbToDTO(req *ad_v1.RejectAdRequest) dto.RejectAdInput {
	adID, _ := uuid.Parse(req.GetId())
	return dto.RejectAdInput{
		AdID: adID,
	}
}

func MapRejectAdDTOToPb(out dto.RejectAdOutput) *ad_v1.RejectAdResponse {
	return &ad_v1.RejectAdResponse{Success: out.Success}
}

func MapDeleteAdPbToDTO(req *ad_v1.DeleteAdRequest, sellerID uuid.UUID) dto.DeleteAdInput {
	adID, _ := uuid.Parse(req.GetId())
	return dto.DeleteAdInput{
		AdID:     adID,
		SellerID: sellerID,
	}
}

func MapDeleteAdDTOToPb(out dto.DeleteAdOutput) *ad_v1.DeleteAdResponse {
	return &ad_v1.DeleteAdResponse{Success: out.Success}
}

func MapDeleteAllAdsPbToDTO(req *ad_v1.DeleteAllAdsRequest) dto.DeleteAllAdsInput {
	sellerID, _ := uuid.Parse(req.GetSellerId())
	return dto.DeleteAllAdsInput{SellerID: sellerID}
}

func MapDeleteAllAdsDTOToPb(out dto.DeleteAllAdsOutput) *ad_v1.DeleteAllAdsResponse {
	return &ad_v1.DeleteAllAdsResponse{Success: out.Success}
}

func MapListAdsPbToDTO(_ *ad_v1.ListAdsRequest, sellerID uuid.UUID) dto.ListSellerAdsInput {
	return dto.ListSellerAdsInput{SellerID: sellerID}
}

func MapListAdsDTOToPb(out dto.ListSellerAdsOutput) *ad_v1.ListAdsResponse {
	mapped := make([]*ad_v1.Ad, len(out.Ads))
	for i := range mapped {
		mapped[i] = mapAdDTOToPB(out.Ads[i])
	}

	return &ad_v1.ListAdsResponse{Ads: mapped}
}

func MapListAllAdsPbToDTO(req *ad_v1.ListAllAdsRequest) dto.ListAllAdsInput {
	return dto.ListAllAdsInput{
		Limit:  int(req.GetLimit()),
		Offset: int(req.GetOffset()),
	}
}

func MapListAllAdsDTOToPb(out dto.ListAllAdsOutput) *ad_v1.ListAllAdsResponse {
	mapped := make([]*ad_v1.Ad, len(out.Ads))
	for i := range mapped {
		mapped[i] = mapAdDTOToPB(out.Ads[i])
	}

	return &ad_v1.ListAllAdsResponse{Ads: mapped}
}
