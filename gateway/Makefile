.PHONY: proto gqlgen generate run tidy

# Regenerate Go protobuf/grpc stubs for all three downstream services.
# Assumes each service's .proto lives under proto/<name>/ as copied into
# this repo, and that each service's own go_package option is already
# correct (it is, in the files you gave me).
proto:
	protoc \
		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		-I proto \
		proto/auth/authservice.proto \
		proto/user/userservice.proto \
		proto/ad/adservice.proto

# Regenerate gqlgen's exec/model code from graph/schema.graphqls +
# gqlgen.yml. Safe to re-run any time the schema changes; it will not
# clobber the resolver bodies in graph/schema.resolvers.go, only fill in
# methods that are missing.
gqlgen:
	go run github.com/99designs/gqlgen generate

generate: proto gqlgen

tidy:
	go mod tidy

run:
	go run ./cmd/gateway
