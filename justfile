_default:
  @just --list --list-heading '' --list-prefix '' --unsorted

# Build the provider
@build:
  go build -o terraform-provider-test

# Install the provider locally
@install: build
  mkdir -p ~/.terraform.d/plugins/registry.terraform.io/jd-ucpa/test/0.1.0/darwin_arm64
  mv terraform-provider-test ~/.terraform.d/plugins/registry.terraform.io/jd-ucpa/test/0.1.0/darwin_arm64

# Run tests
@test:
  go test -i ./... || exit 1
  go test ./... -timeout=30s -parallel=4

# Run acceptance tests
@testacc:
  TF_ACC=1 go test ./... -v -timeout 120m
