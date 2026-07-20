package phone

import (
	"context"

	pkgerrs "github.com/maket12/ads-service/userservice/pkg/errs"

	"github.com/nyaruka/phonenumbers"
)

type Validator struct {
	defaultRegion string // set to empty string for international support
}

func NewValidator(defaultRegion string) *Validator {
	return &Validator{defaultRegion: defaultRegion}
}

func (v *Validator) Validate(_ context.Context, phone string) (string, error) {
	num, err := phonenumbers.Parse(phone, v.defaultRegion)
	if err != nil {
		return "", pkgerrs.NewValueInvalidErrorWithReason("phone", err)
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", pkgerrs.NewValueInvalidError("phone")
	}

	return phonenumbers.Format(num, phonenumbers.E164), nil
}
