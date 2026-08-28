package graph

import (
	"github.com/maket12/ads-service/backend/adservice/api/proto/generated/ad_v1"
	authutils "github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/maket12/ads-service/backend/gateway/graph/model"
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

func mapFloatPtrToInt(ptr *float64) *int64 {
	if ptr == nil {
		return nil
	}
	return authutils.VPtr(int64(*ptr))
}

func mapAdGRPCToGateway(ad *ad_v1.Ad) *model.Ad {
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

func mapAdListGRPCToGateway(ads []*ad_v1.Ad) []*model.Ad {
	mapped := make([]*model.Ad, len(ads))
	for i := range mapped {
		mapped[i] = mapAdGRPCToGateway(ads[i])
	}
	return mapped
}
