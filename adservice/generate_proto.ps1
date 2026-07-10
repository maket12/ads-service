$ErrorActionPreference = 'Stop'

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$repoRoot = Resolve-Path (Join-Path $scriptDir '..')
$protoDir = Join-Path $scriptDir 'api/proto'

if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    throw 'protoc is not installed or not available in PATH'
}

$goBin = Join-Path (go env GOPATH) 'bin'
$env:Path = "$goBin;$env:Path"

if (-not (Get-Command protoc-gen-go -ErrorAction SilentlyContinue)) {
    Write-Host 'Installing protoc-gen-go...'
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
}

if (-not (Get-Command protoc-gen-go-grpc -ErrorAction SilentlyContinue)) {
    Write-Host 'Installing protoc-gen-go-grpc...'
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
}

$modulePath = go list -m
$protoFiles = Get-ChildItem -Path $protoDir -Filter '*.proto' | ForEach-Object { $_.FullName }

if ($protoFiles.Count -eq 0) {
    throw "No .proto files found in $protoDir"
}

Write-Host 'Generating protobuf code for adservice...'
& protoc --proto_path=$protoDir --proto_path=$repoRoot --go_out=$repoRoot --go_opt=module=$modulePath --go-grpc_out=$repoRoot --go-grpc_opt=module=$modulePath $protoFiles
Write-Host 'Done.'
