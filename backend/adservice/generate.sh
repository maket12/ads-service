#!/usr/bin/env bash
set -euo pipefail

# Place this script directly in the microservice root, e.g.:
#   authservice/generate.sh
# Run it from the microservice root:
#   ./generate.sh
#
# Expected structure:
#   authservice/
#     api/proto/*.proto
#     pkg/generated/...   (generated code goes here, path taken from go_package minus module prefix)

# Directory where the script itself is located == microservice root
SERVICE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Go module path, must match the "module" line in the root go.mod
MODULE_NAME="github.com/maket12/ads-service"

# Root of the go module (repo root), one level above the microservice
MODULE_ROOT="$(cd "${SERVICE_ROOT}/.." && pwd)"

PROTO_DIR="${SERVICE_ROOT}/api/proto"
OUT_DIR="${MODULE_ROOT}"

# Check required tools
command -v protoc >/dev/null 2>&1 || { echo "protoc not found. Install protobuf-compiler."; exit 1; }
command -v protoc-gen-go >/dev/null 2>&1 || { echo "protoc-gen-go not found. Run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"; exit 1; }
command -v protoc-gen-go-grpc >/dev/null 2>&1 || { echo "protoc-gen-go-grpc not found. Run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"; exit 1; }

if [ ! -d "${PROTO_DIR}" ]; then
  echo "Proto directory not found: ${PROTO_DIR}"
  exit 1
fi

echo "Generating protobuf/grpc code for $(basename "${SERVICE_ROOT}")..."
echo "  proto dir: ${PROTO_DIR}"
echo "  out dir:   ${OUT_DIR}"

find "${PROTO_DIR}" -name '*.proto' -print0 | while IFS= read -r -d '' proto_file; do
  echo "  -> ${proto_file}"
  protoc \
    --proto_path="${PROTO_DIR}" \
    --go_out="${OUT_DIR}" \
    --go_opt=module="${MODULE_NAME}" \
    --go-grpc_out="${OUT_DIR}" \
    --go-grpc_opt=module="${MODULE_NAME}" \
    "${proto_file}"
done

echo "Done."