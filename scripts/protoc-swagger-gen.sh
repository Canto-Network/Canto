#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./third_party
cd third_party
git clone --branch v0.50.6 --single-branch --depth 1 https://github.com/cosmos/cosmos-sdk
git clone --branch v8.2.1 --single-branch --depth 1 https://github.com/cosmos/ibc-go
git clone --branch dudong2/feat/cosmos-sdk@v0.50.x-cometbft@v0.38.0-2 --single-branch --depth 1 https://github.com/b-harvest/ethermint
cd ..

mkdir -p ./tmp-swagger-gen

proto_dirs=$(find ./proto ./third_party/cosmos-sdk/proto ./third_party/ibc-go/proto ./third_party/ethermint/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  # TODO: migrate to `buf build`
  if [[ ! -z "$query_file" ]]; then
    buf generate --template ./proto/buf.gen.swagger.yaml $query_file
  fi
done

# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./third_party
rm -rf ./tmp-swagger-gen

# generate binary for static server
#statik -src=./client/docs/swagger-ui -dest=./client/docs
