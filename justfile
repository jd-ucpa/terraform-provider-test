_default:
  @just --list --list-heading '' --list-prefix '' --unsorted

# liste les 'namespaces' de tests
ls:
  cat test/*.go | grep '^func.*testing.T' | sed -E 's/^func (Test.*)_.*/\1/' | sort -u

# - just test (tous les tests)
# - just test TestAccPro (tous les tests dont le nom contient TestAccPro)
# Exécute les tests d'acceptance (tous ou une partie)
test *name:
  #!/usr/bin/env bash
  if [[ -z "{{name}}" ]]; then
    TF_ACC=1 go test -v ./test -timeout 120m
  else
    TF_ACC=1 go test -v ./test -run "{{name}}" -timeout 120m
  fi

# installe ou update localement tfplugindocs (si nécessaire)
@down:
  ./bin/down.sh

# regénère la documentation
@docs:
  rm -rf docs/
  ./bin/tfplugindocs generate --website-source-dir templates

# Build + install le provider localement
@install: 
  go build -ldflags="-s -w" -o terraform-provider-test
  mkdir -p ~/.terraform.d/plugins/registry.terraform.io/jd-ucpa/test/$(cat internal/VERSION)/darwin_arm64
  mv terraform-provider-test ~/.terraform.d/plugins/registry.terraform.io/jd-ucpa/test/$(cat internal/VERSION)/darwin_arm64
  cp terraform.rc ~/.terraform.d

# init example/
@init:
  export TF_PLUGIN_CACHE_DIR=$HOME/.terraform.d/plugin-cache && export AWS_PROFILE=9539 && cd example && sed -i '' '/provider "registry.terraform.io\/jd-ucpa\/test" {/,/}/d' .terraform.lock.hcl && terraform init && terraform plan

# apply example/
@apply:
  export AWS_PROFILE=9539 && cd example && terraform apply -auto-approve

# install init apply
@iia: install init apply
