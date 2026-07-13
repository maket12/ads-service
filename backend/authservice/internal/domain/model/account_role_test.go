package model_test

import (
	"strings"
	"testing"

	pkgerrs "github.com/maket12/ads-service/authservice/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountRole(t *testing.T) {
	type testCase struct {
		name      string
		accountID uuid.UUID
		expect    error
	}

	var tests = []testCase{
		{
			name:      "success",
			accountID: uuid.New(),
			expect:    nil,
		},
		{
			name:      "nullable account id",
			accountID: uuid.Nil,
			expect:    pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accRole, err := NewAccountRole(tt.accountID)
			if tt.expect == nil {
				require.NoError(t, err)
				require.NotNil(t, accRole)
				assert.Equal(t, tt.accountID, accRole.AccountID())
				assert.Equal(t, RoleUser, accRole.Role())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, pkgerrs.ErrValueIsInvalid)
				assert.Nil(t, accRole)
			}
		})
	}
}

func TestAccountRole_Assign(t *testing.T) {
	type testCase struct {
		name   string
		role   string
		expect error
	}

	var tests = []testCase{
		{
			name:   "success - admin",
			role:   "admin",
			expect: nil,
		},
		{
			name:   "success - user",
			role:   "user",
			expect: nil,
		},
		{
			name:   "success - in upper case",
			role:   "ADMIN",
			expect: nil,
		},
		{
			name:   "invalid role value",
			role:   "unknown",
			expect: pkgerrs.ErrValueIsInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accRole, _ := NewAccountRole(uuid.New())

			err := accRole.Assign(tt.role)

			if tt.expect == nil {
				require.NoError(t, err)
				assert.Equal(t, Role(strings.ToLower(tt.role)), accRole.Role())
			} else {
				require.Error(t, err)
				assert.ErrorIs(t, err, pkgerrs.ErrValueIsInvalid)
				assert.NotEqual(t, Role(tt.role), accRole.Role())
			}
		})
	}
}
