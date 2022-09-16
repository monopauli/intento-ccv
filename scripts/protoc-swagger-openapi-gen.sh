#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen
proto_dirs=$(find ./proto ./third_party/proto/cosmos ./third_party/proto/ibc -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    # 1. Get buf from https://github.com/bufbuild/buf/releases/tag/v1.0.0-rc12
    # Note that v1.0.0-rc12 is the last version with the "buf protoc" subcommand
    # 2. Get swagger protoc plugin with `go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0`

    buf protoc  \
    -I "proto" \
    -I "third_party/proto" \
    "$query_file" \
    --swagger_out=./tmp-swagger-gen \
    --swagger_opt=logtostderr=true \
    --swagger_opt=fqn_for_swagger_name=true \
    --swagger_opt=simple_operation_ids=true
  fi
done

# service.swagger.json doesn't work for some reasone
(
  cd ./client/docs

  # Generate config.json
  # There's some operationIds naming collision, for sake of automation we're
  # giving all of them a unique name
  find ../../tmp-swagger-gen -name 'query.swagger.json' | 
    sort |
    awk '{print "{\"url\":\""$1"\",\"operationIds\":{\"rename\":{\"Params\":\""$1"Params\",\"DelegatorValidators\":\""$1"DelegatorValidators\",\"UpgradedConsensusState\":\""$1"UpgradedConsensusState\"}}}"}' |
    jq -s '{swagger:"2.0","info":{"title":"Secret Network - gRPC Gateway docs","description":"A REST interface for queries and transactions","version":"'"$(git describe --tags $(git rev-list --tags --max-count=1) | perl -pe 's/-(beta|alpha).*//')"'"},apis:.} | .apis += [{"url":"./swagger_legacy.yaml","dereference":{"circular":"ignore"}}]' > ./config.json

  # Derive openapi & swagger from config.json
  yarn install
  yarn combine
  yarn convert
  yarn build
)

# clean swagger files
rm -rf ./tmp-swagger-gen