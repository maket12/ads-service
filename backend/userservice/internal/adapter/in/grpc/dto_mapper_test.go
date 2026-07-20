package grpc_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/maket12/ads-service/userservice/internal/adapter/in/grpc"
	"github.com/maket12/ads-service/userservice/internal/app/dto"
	"github.com/maket12/ads-service/userservice/pkg/generated/user_v1"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapGetProfileDTOToPb(t *testing.T) {
	accountID := uuid.New()
	fName := gofakeit.FirstName()
	lName := gofakeit.LastName()
	phone := gofakeit.PhoneFormatted()
	avatar := gofakeit.URL()
	bio := gofakeit.Bio()
	updatedAt := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)

	type testCase struct {
		name  string
		input dto.GetProfileOutput
		want  *user_v1.GetProfileResponse
	}

	testCases := []testCase{
		{
			name: "maps all fields correctly",
			input: dto.GetProfileOutput{
				AccountID: accountID,
				FirstName: &fName,
				LastName:  &lName,
				Phone:     &phone,
				AvatarURL: &avatar,
				Bio:       &bio,
				UpdatedAt: updatedAt,
			},
			want: &user_v1.GetProfileResponse{
				AccountId: accountID.String(),
				FirstName: &fName,
				LastName:  &lName,
				Phone:     &phone,
				AvatarUrl: &avatar,
				Bio:       &bio,
				UpdatedAt: timestamppb.New(updatedAt),
			},
		},
		{
			name: "maps zero UUID and zero time",
			input: dto.GetProfileOutput{
				AccountID: uuid.Nil,
				UpdatedAt: time.Time{},
			},
			want: &user_v1.GetProfileResponse{
				AccountId: uuid.Nil.String(),
				UpdatedAt: timestamppb.New(time.Time{}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := grpc.MapGetProfileDTOToPb(tc.input)

			if !reflect.DeepEqual(got.AccountId, tc.want.AccountId) ||
				!reflect.DeepEqual(got.FirstName, tc.want.FirstName) ||
				!reflect.DeepEqual(got.LastName, tc.want.LastName) ||
				!reflect.DeepEqual(got.Phone, tc.want.Phone) ||
				!reflect.DeepEqual(got.AvatarUrl, tc.want.AvatarUrl) ||
				!reflect.DeepEqual(got.Bio, tc.want.Bio) {
				t.Errorf("MapGetProfileDTOToPb() scalar fields mismatch, got = %+v, want = %+v", got, tc.want)
			}

			// timestamppb.Timestamp carries internal state (e.g. sizeCache) that
			// reflect.DeepEqual on the raw struct can flag as different even when
			// the represented instant is identical, so compare via AsTime().
			if !got.UpdatedAt.AsTime().Equal(tc.want.UpdatedAt.AsTime()) {
				t.Errorf("MapGetProfileDTOToPb() UpdatedAt = %v, want %v", got.UpdatedAt.AsTime(), tc.want.UpdatedAt.AsTime())
			}
		})
	}
}

func TestMapUpdateProfilePbToDTO(t *testing.T) {
	accountID := uuid.New()
	fName := gofakeit.FirstName()
	lName := gofakeit.LastName()
	phone := gofakeit.PhoneFormatted()
	avatar := gofakeit.URL()
	bio := gofakeit.Bio()

	type testCase struct {
		name      string
		accountID uuid.UUID
		req       *user_v1.UpdateProfileRequest
		want      dto.UpdateProfileInput
	}

	testCases := []testCase{
		{
			name:      "maps all fields correctly",
			accountID: accountID,
			req: &user_v1.UpdateProfileRequest{
				FirstName: &fName,
				LastName:  &lName,
				Phone:     &phone,
				AvatarUrl: &avatar,
				Bio:       &bio,
			},
			want: dto.UpdateProfileInput{
				AccountID: accountID,
				FirstName: &fName,
				LastName:  &lName,
				Phone:     &phone,
				AvatarURL: &avatar,
				Bio:       &bio,
			},
		},
		{
			name:      "maps zero account id",
			accountID: uuid.Nil,
			req:       &user_v1.UpdateProfileRequest{},
			want:      dto.UpdateProfileInput{AccountID: uuid.Nil},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := grpc.MapUpdateProfilePbToDTO(tc.accountID, tc.req)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("MapUpdateProfilePbToDTO() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestMapUpdateProfileDTOToPb(t *testing.T) {
	type testCase struct {
		name  string
		input dto.UpdateProfileOutput
		want  *user_v1.UpdateProfileResponse
	}

	testCases := []testCase{
		{
			name:  "maps success true",
			input: dto.UpdateProfileOutput{Success: true},
			want:  &user_v1.UpdateProfileResponse{Success: true},
		},
		{
			name:  "maps success false",
			input: dto.UpdateProfileOutput{Success: false},
			want:  &user_v1.UpdateProfileResponse{Success: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := grpc.MapUpdateProfileDTOToPb(tc.input)

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("MapUpdateProfileDTOToPb() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
