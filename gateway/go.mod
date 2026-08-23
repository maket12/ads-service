module github.com/maket12/ads-service/gateway

go 1.22

require (
	github.com/99designs/gqlgen v0.17.49
	github.com/vektah/gqlparser/v2 v2.5.16
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2

	// Sibling services' generated protobuf/grpc packages.
	// In a monorepo these resolve normally once go.work (or a replace
	// directive below) points at their local paths. See README.md.
	github.com/maket12/ads-service/adservice v0.0.0
	github.com/maket12/ads-service/authservice v0.0.0
	github.com/maket12/ads-service/userservice v0.0.0
)

// Adjust these paths to wherever the sibling services actually live
// relative to this gateway module, or delete them if you're using a
// go.work file instead (recommended for a monorepo). See README.md.
replace (
	github.com/maket12/ads-service/adservice => ../adservice
	github.com/maket12/ads-service/authservice => ../authservice
	github.com/maket12/ads-service/userservice => ../userservice
)
