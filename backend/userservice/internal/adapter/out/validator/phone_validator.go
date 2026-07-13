package validator

import (
	"context"

	pkgerrs "github.com/maket12/ads-service/pkg/errs"

	"github.com/nyaruka/phonenumbers"
)

type PhoneValidator struct {
	// Set to empty string for international support
	defaultRegion string
}

func NewPhoneValidator(defaultRegion string) *PhoneValidator {
	return &PhoneValidator{defaultRegion: defaultRegion}
}

func (v *PhoneValidator) Validate(_ context.Context, phone string) (string, error) {
	num, err := phonenumbers.Parse(phone, v.defaultRegion)
	if err != nil {
		return "", pkgerrs.NewValueInvalidErrorWithReason("phone", err)
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", pkgerrs.NewValueInvalidError("phone")
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}
