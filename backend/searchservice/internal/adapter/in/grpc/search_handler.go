package grpc

import (
	"context"
	"log/slog"

	"github.com/maket12/ads-service/backend/authservice/pkg/utils"
	"github.com/maket12/ads-service/backend/searchservice/api/proto/generated/search_v1"
	"github.com/maket12/ads-service/backend/searchservice/internal/app/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SearchHandler struct {
	search_v1.UnimplementedSearchServiceServer
	log         *slog.Logger
	searchAdsUC *usecase.SearchAdsUC
}

func NewSearchHandler(log *slog.Logger, searchAdsUC *usecase.SearchAdsUC) *SearchHandler {
	return &SearchHandler{
		log:         log,
		searchAdsUC: searchAdsUC,
	}
}

func (h *SearchHandler) SearchAds(ctx context.Context, req *search_v1.SearchAdsRequest) (*search_v1.SearchAdsResponse, error) {
	if err := h.authorize(ctx); err != nil {
		return nil, err
	}

	ucResp, err := h.searchAdsUC.Execute(ctx, MapSearchAdsPbToDTO(req))
	if err != nil {
		code, msg := h.handleError(ctx, err, "failed to search ads")
		return nil, status.Error(code, msg)
	}

	return MapSearchAdsDTOToPb(ucResp), nil
}

func (h *SearchHandler) authorize(ctx context.Context) error {
	if _, err := utils.ExtractAccountID(ctx); err != nil {
		outErr := gRPCError(err)
		return status.Error(outErr.Code, outErr.Message)
	}
	return nil
}

func (h *SearchHandler) handleError(
	ctx context.Context, err error,
	logMsg string,
) (codes.Code, string) {
	outErr := gRPCError(err)
	h.log.LogAttrs(ctx, outErr.Level, logMsg,
		slog.Int("code", int(outErr.Code)),
		slog.String("public_msg", outErr.Message),
		slog.Any("reason", outErr.Reason),
	)
	return outErr.Code, outErr.Message
}
