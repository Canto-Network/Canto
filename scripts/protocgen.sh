#!/usr/bin/env bash


set -eo pipefail


protoc_gen_gocosmos() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the cosmos-sdk folder."
    return 1
  fi

  go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest 2>/dev/null
}

protoc_gen_doc() {
  go get github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc 2>/dev/null
}

echo "1"

protoc_gen_gocosmos
#protoc_gen_doc

echo "2"

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  buf protoc \
    -I "proto" \
    -I "third_party/proto" \
    --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
    --grpc-gateway_out=logtostderr=true,allow_colon_final_segments=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')

done

echo "3"

# command to generate docs using protoc-gen-doc
# TODO: migrate to `buf build`
# buf alpha protoc \
# -I "proto" \
# -I "third_party/proto" \
# --doc_out=./docs/protocol \
# --doc_opt=./docs/protodoc-markdown.tmpl,proto-docs.md \
# $(find "$(pwd)/proto" -maxdepth 5 -name '*.proto')


# move proto files to the right places
cp -r github.com/Canto-Network/Canto/v7/x/* x/
rm -rf github.com
