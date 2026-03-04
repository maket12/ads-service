package validator_test

import (
	"context"
	"github.com/maket12/ads-service/userservice/internal/adapter/out/validator"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPhoneValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		region string
	}

	var tests = []testCase{
		{
			name:   "success - russia",
			region: "RU",
		},
		{
			name:   "success - china",
			region: "CN",
		},
		{
			name:   "success - region not specified",
			region: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := validator.NewPhoneValidator(tt.region)
			assert.NotNil(t, val)
		})
	}
}

func TestPhoneValidator_Validate(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name    string
		phone   string
		expect  string
		wantErr bool
	}

	var tests = []testCase{
		{
			name:    "success/1",
			phone:   "+7 (999) 123-45-13", // Russia
			expect:  "+79991234513",
			wantErr: false,
		},
		{
			name:    "success/2",
			phone:   "+1-202-456-1111", // USA
			expect:  "+12024561111",
			wantErr: false,
		},
		{
			name:    "success/3",
			phone:   "+79139589457",
			expect:  "+79139589457",
			wantErr: false,
		},
		{
			name:    "failure/1",
			phone:   "781425",
			expect:  "",
			wantErr: true,
		},
		{
			name:    "failure/2",
			phone:   "+09141785",
			expect:  "",
			wantErr: true,
		},
		{
			name:    "failure/3",
			phone:   "+79139549897241514125",
			expect:  "",
			wantErr: true,
		},
	}

	phoneValidator := validator.NewPhoneValidator("")
	testCtx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normPhone, err := phoneValidator.Validate(testCtx, tt.phone)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, normPhone)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expect, normPhone)
			}
		})
	}
}
