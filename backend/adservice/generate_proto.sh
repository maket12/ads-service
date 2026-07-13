#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PROTO_DIR="$SCRIPT_DIR/api/proto"

if ! command -v protoc >/dev/null 2>&1; then
  echo "protoc is not installed or not available in PATH" >&2
  exit 1
fi

if ! command -v protoc-gen-go >/dev/null 2>&1; then
  echo "protoc-gen-go is not installed or not available in PATH" >&2
  exit 1
fi

if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
  echo "protoc-gen-go-grpc is not installed or not available in PATH" >&2
  exit 1
fi

MODULE_PATH="$(cd "$REPO_ROOT" && go list -m)"
PROTO_FILES=("$PROTO_DIR"/*.proto)

if [ ${#PROTO_FILES[@]} -eq 0 ]; then
  echo "No .proto files found in $PROTO_DIR" >&2
  exit 1
fi

echo "Generating protobuf code for adservice..."
protoc \
  --proto_path="$PROTO_DIR" \
  --proto_path="$REPO_ROOT" \
  --go_out="$REPO_ROOT" \
  --go_opt=module="$MODULE_PATH" \
  --go-grpc_out="$REPO_ROOT" \
  --go-grpc_opt=module="$MODULE_PATH" \
  "${PROTO_FILES[@]}"

echo "Done."
