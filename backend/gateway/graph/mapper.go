package graph

import (
	"github.com/maket12/ads-service/backend/adservice/api/proto/generated/ad_v1"
	authutils "github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/maket12/ads-service/backend/gateway/graph/model"
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	"github.com/maket12/ads-service/backend/userservice/api/proto/generated/user_v1"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "INVALID_ARGUMENT"}}
	case codes.NotFound:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "NOT_FOUND"}}
	case codes.AlreadyExists:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "ALREADY_EXISTS"}}
	case codes.PermissionDenied:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "FORBIDDEN"}}
	case codes.FailedPrecondition:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "FAILED_PRECONDITION"}}
	case codes.Unauthenticated:
		return &gqlerror.Error{Message: st.Message(), Extensions: map[string]interface{}{"code": "UNAUTHENTICATED"}}
	default:
		return &gqlerror.Error{Message: "internal error", Extensions: map[string]interface{}{"code": "INTERNAL"}}
	}
}

func mapFloatPtrToIntPtr(ptr *float64) *int64 {
	if ptr == nil {
		return nil
	}
	return authutils.VPtr(int64(*ptr))
}

func mapIntPtrToInt32(ptr *int) int32 {
	if ptr == nil {
		return 0
	}
	return int32(*ptr)
}

func mapStringPtrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func MapUserGRPCToGateway(user *user_v1.GetProfileResponse, role string) *model.User {
	if user == nil {
		return nil
	}
	return &model.User{
		ID:        user.AccountId,
		Role:      model.UserRole(role),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		AvatarURL: user.AvatarUrl,
		Bio:       user.Bio,
		UpdatedAt: authutils.VPtr(user.UpdatedAt.String()),
	}
}

func MapAdGRPCToGateway(ad *ad_v1.Ad) *model.Ad {
	if ad == nil {
		return nil
	}
	return &model.Ad{
		ID:          ad.Id,
		SellerID:    ad.SellerId,
		Title:       ad.Title,
		Description: ad.Description,
		Price:       float64(ad.Price),
		Category:    ad.Category,
		Status:      model.AdStatus(ad.Status),
		Images:      ad.Images,
		CreatedAt:   ad.CreatedAt.String(),
		UpdatedAt:   authutils.VPtr(ad.UpdatedAt.String()),
	}
}

func MapAdListGRPCToGateway(ads []*ad_v1.Ad) []*model.Ad {
	mapped := make([]*model.Ad, len(ads))
	for i := range mapped {
		mapped[i] = MapAdGRPCToGateway(ads[i])
	}
	return mapped
}

func MapSearchInputToRequest(in *model.SearchInput) *search_v1.SearchAdsRequest {
	return &search_v1.SearchAdsRequest{
		Text:      mapStringPtrToString(in.Text),
		Category:  in.Category,
		PriceFrom: mapFloatPtrToIntPtr(in.PriceFrom),
		PriceTo:   mapFloatPtrToIntPtr(in.PriceTo),
		Limit:     mapIntPtrToInt32(in.Limit),
		Offset:    mapIntPtrToInt32(in.Offset),
		SortBy:    mapStringPtrToString(in.SortBy),
	}
}

func mapAdIndexGRPCToGateway(adIdx *search_v1.AdIndex) *model.AdIndex {
	if adIdx == nil {
		return nil
	}
	return &model.AdIndex{
		ID:        adIdx.Id,
		Title:     adIdx.Title,
		Price:     float64(adIdx.Price),
		Category:  adIdx.Category,
		MainImage: adIdx.MainImage,
	}
}

func MapAdIndexListGRPCToGateway(adIndexes []*search_v1.AdIndex) []*model.AdIndex {
	mapped := make([]*model.AdIndex, len(adIndexes))
	for i := range mapped {
		mapped[i] = mapAdIndexGRPCToGateway(adIndexes[i])
	}
	return mapped
}
