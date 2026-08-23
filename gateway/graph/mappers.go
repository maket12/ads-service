package graph

import (
	"time"

	"github.com/99designs/gqlgen/graphql"
	"google.golang.org/protobuf/types/known/timestamppb"

	ad_v1 "github.com/maket12/ads-service/adservice/pkg/generated/ad_v1"

	"github.com/maket12/ads-service/gateway/graph/model"
)

// toModelAd converts an AdService Ad message into the GraphQL Ad model.
func toModelAd(a *ad_v1.Ad) *model.Ad {
	if a == nil {
		return nil
	}
	return &model.Ad{
		AdID:        graphql.ID(a.GetAdId()),
		SellerID:    graphql.ID(a.GetSellerId()),
		Title:       a.GetTitle(),
		Description: optStr(a.Description),
		Price:       float64(a.GetPrice()),
		Status:      model.AdStatus(a.GetStatus()),
		Images:      refAll(a.GetImages()),
		CreatedAt:   timestampPtr(a.GetCreatedAt()),
		UpdatedAt:   timestampPtr(a.GetUpdatedAt()),
	}
}

// optStr passes an already-optional proto3 `optional string` field through
// unchanged (proto3 optional scalars are already *string on the Go side).
func optStr(s *string) *string { return s }

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// timestampPtr renders a google.protobuf.Timestamp as RFC3339, matching the
// plain `String` scalar used for createdAt/updatedAt in the schema.
func timestampPtr(ts *timestamppb.Timestamp) *string {
	if ts == nil {
		return nil
	}
	s := ts.AsTime().Format(time.RFC3339)
	return &s
}

// derefAll converts the []*string gqlgen generates for a GraphQL [String]!
// argument into the []string the proto messages expect. Nil elements (an
// explicit `null` inside the list) are dropped.
func derefAll(in []*string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != nil {
			out = append(out, *s)
		}
	}
	return out
}

// refAll converts a []string from a proto message into the []*string
// gqlgen expects for a [String]! field.
func refAll(in []string) []*string {
	out := make([]*string, len(in))
	for i := range in {
		out[i] = &in[i]
	}
	return out
}
