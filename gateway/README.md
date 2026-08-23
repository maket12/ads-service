# gateway

GraphQL gateway (gqlgen) in front of `authservice`, `userservice`, and
`adservice` over gRPC.

## Why two `go generate` steps are needed before this builds

I wrote every hand-authorable file (resolvers, gRPC client wiring, auth
middleware, models, main.go). Two pieces are machine-generated from your
`.proto` files and `schema.graphqls`, and I don't have `protoc` or `go`
available in this sandbox to run them for you:

1. **Protobuf/gRPC stubs** (`*.pb.go`, `*_grpc.pb.go`) for the three
   services — needed by `internal/clients` and the resolvers.
2. **gqlgen's exec engine** (`graph/generated/generated.go`) — the
   reflection-heavy glue that dispatches incoming GraphQL fields to your
   resolver methods.

I copied `graph/model/models_gen.go` by hand to match exactly what gqlgen
would produce, and `graph/schema.resolvers.go` uses the exact method
signatures gqlgen v0.17 generates for this schema — so once you run the
commands below, everything should slot together without edits.

## Setup

1. **Arrange the monorepo layout.** This gateway expects to sit alongside
   your three services and import their generated packages directly, e.g.:

   ```
   ads-service/
     gateway/        <- this module
     authservice/
     userservice/
     adservice/
   ```

   Easiest: add a `go.work` at the repo root:

   ```
   go work init ./gateway ./authservice ./userservice ./adservice
   ```

   and delete the `replace` block in `gateway/go.mod` (go.work supersedes
   it). Otherwise, adjust the `replace` paths in `go.mod` to match your
   actual layout.

2. **Install the codegen tools** (once):

   ```
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

   (`protoc` itself via your OS package manager — `brew install protobuf`,
   `apt install protobuf-compiler`, etc.)

3. **Generate everything:**

   ```
   cd gateway
   make generate   # = make proto + make gqlgen
   go mod tidy
   make run
   ```

   `make proto` regenerates the stubs from `proto/*/*.proto` — the same
   `.proto` files you gave me, copied in under `proto/`. If your actual
   `authservice`/`userservice`/`adservice` repos already generate their own
   stubs elsewhere (matching their `go_package` options), point
   `internal/clients/clients.go`'s imports at those instead and skip
   `make proto` — `go_package` in your `.proto` files already points at
   `.../pkg/generated/<name>_v1`, so this only matters if you'd rather
   generate centrally from the gateway.

   `make gqlgen` runs `go run github.com/99designs/gqlgen generate`, which
   creates `graph/generated/generated.go` from `graph/schema.graphqls` +
   `gqlgen.yml`.

4. Configure downstream addresses via env vars (defaults shown):

   ```
   AUTH_SERVICE_ADDR=localhost:9091
   USER_SERVICE_ADDR=localhost:9092
   AD_SERVICE_ADDR=localhost:9093
   GATEWAY_ADDR=:8080
   ```

5. Open `http://localhost:8080/` for the GraphQL playground.

## How auth flows through the gateway

None of the downstream RPCs that need "who is calling" (`GetProfile`,
`UpdateProfile`, `CreateAd`, `UpdateAd`, ...) take an `account_id` field —
that's by design, since only the gateway talks to `AuthService`.

- `internal/middleware.Auth` reads `Authorization: Bearer <token>` on every
  HTTP request, calls `AuthService.ValidateAccessToken`, and — if valid —
  stashes `{account_id, role}` on the request context. Missing/invalid
  tokens are treated as anonymous, not rejected outright, so public queries
  (e.g. `ad(adId: ...)`) still work without a token.
- Resolvers that require a caller call `authctx.MustFromContext(ctx)` and
  return `authctx.ErrUnauthenticated` if there isn't one.
- Right before any downstream gRPC call, resolvers wrap the context with
  `authctx.Outgoing(ctx)`, which attaches `x-account-id` / `x-role` gRPC
  metadata headers. Your `userservice`/`adservice` implementations should
  read those headers (interceptor or per-handler) instead of trusting a
  client-supplied id — this is the standard "gateway pre-authenticates,
  services trust internal metadata" pattern, and matches why
  `GetProfileRequest`/`CreateAdRequest` don't carry an id themselves.

## Known TODOs left in the code

- `AssignRole` checks `caller.Role != "admin"` as a placeholder — wire up
  your real role/permission model.
- `internal/clients/clients.go` dials with insecure credentials — swap in
  TLS creds before this leaves a trusted network.
- `Ad.price` / proto `price` is treated as a plain unit conversion
  (`int64` ⇄ `float64`) — adjust in `graph/mappers.go` if it's actually
  minor units (cents) that need dividing/multiplying by 100.
- `listAds` / `listAllAds` / `deleteAllAds` RPCs exist on `AdService` but
  aren't exposed in `schema.graphqls` yet — add `Query`/`Mutation` fields
  and resolver cases for them if you need those.
