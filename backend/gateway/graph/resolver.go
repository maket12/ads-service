package graph

import (
	"github.com/maket12/ads-service/backend/adservice/api/proto/generated/ad_v1"
	"github.com/maket12/ads-service/backend/authservice/api/proto/generated/auth_v2"
	"github.com/maket12/ads-service/backend/userservice/api/proto/generated/user_v1"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AuthClient auth_v2.AuthServiceClient
	UserClient user_v1.UserServiceClient
	AdClient   ad_v1.AdServiceClient
}
