package model_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/authservice/internal/domain/model"
	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerificationToken(t *testing.T) {
	type testCase struct {
		name      string
		accountID uuid.UUID
		ttl       time.Duration
		expect    error
	}

	var tests = []testCase{
		{
			name:      "success",
			accountID: uuid.New(),
			ttl:       time.Minute,
			expect:    nil,
		},
		{
			name:      "nullable account id",
			accountID: uuid.Nil,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
		{
			name:      "short ttl",
			accountID: uuid.New(),
			ttl:       time.Second,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()

			vToken, err := model.NewVerificationToken(
				tt.accountID,
				tt.ttl,
			)

			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, vToken)
				assert.NotEmpty(t, vToken.Token())
				assert.Equal(t, tt.accountID, vToken.AccountID())
				assert.Equal(t, tt.ttl, vToken.TTL())
				assert.WithinDuration(t,
					now.Add(tt.ttl),
					vToken.ExpiresAt(),
					time.Millisecond,
				)
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expect)
				assert.Nil(t, vToken)
			}
		})
	}
}

func TestVerificationToken_IsExpired(t *testing.T) {
	vToken := model.RestoreVerificationToken(
		"a-random-token",
		uuid.New(),
		time.Second,
		time.Now().Add(time.Second),
	)
	assert.False(t, vToken.IsExpired())
	time.Sleep(time.Second)
	assert.True(t, vToken.IsExpired())
}
