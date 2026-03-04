package utils_test

import (
	"context"
	"github.com/maket12/ads-service/pkg/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestExtractAccountID(t *testing.T) {
	type testCase struct {
		name    string
		ctx     context.Context
		expect  uuid.UUID
		wantErr bool
	}

	var (
		targetUID = uuid.New()
		tests     = []testCase{
			{
				name: "success",
				ctx: metadata.NewIncomingContext(context.Background(),
					metadata.Pairs("x-account-id", targetUID.String()),
				),
				expect:  targetUID,
				wantErr: false,
			},
			{
				name:    "failure - missing metadata",
				ctx:     context.Background(),
				expect:  uuid.Nil,
				wantErr: true,
			},
			{
				name: "failure - account id is not specified",
				ctx: metadata.NewIncomingContext(context.Background(),
					metadata.Pairs("wants to be", "a digital nomad"),
				),
				expect:  uuid.Nil,
				wantErr: true,
			},
			{
				name: "failure - invalid account id",
				ctx: metadata.NewIncomingContext(context.Background(),
					metadata.Pairs("x-account-id", "not-valid-uuid"),
				),
				expect:  uuid.Nil,
				wantErr: true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, err := utils.ExtractAccountID(tt.ctx)

			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.expect, uid)
		})
	}
}

func TestExtractAccountRole(t *testing.T) {
	type testCase struct {
		name    string
		ctx     context.Context
		expect  string
		wantErr bool
	}

	var (
		targetRole = "admin"
		tests      = []testCase{
			{
				name: "success",
				ctx: metadata.NewIncomingContext(context.Background(),
					metadata.Pairs("x-account-role", targetRole),
				),
				expect:  targetRole,
				wantErr: false,
			},
			{
				name:    "failure - missing metadata",
				ctx:     context.Background(),
				expect:  "",
				wantErr: true,
			},
			{
				name: "failure - account role is not specified",
				ctx: metadata.NewIncomingContext(context.Background(),
					metadata.Pairs("wants to be", "a digital nomad"),
				),
				expect:  "",
				wantErr: true,
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uid, err := utils.ExtractAccountRole(tt.ctx)

			if !tt.wantErr {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			assert.Equal(t, tt.expect, uid)
		})
	}
}
