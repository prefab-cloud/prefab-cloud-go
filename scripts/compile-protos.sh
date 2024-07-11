#!/bin/bash

#run from the root directory

protoc --proto_path=proto-source \
  --go_out=proto --go_opt=paths=source_relative \
  --go_opt=Mproto-source/prefab.proto=github.com/prefab-cloud/prefab-cloud-go/proto \
  proto-source/prefab.proto

# protoc --go_out=. ../prefab-cloud-internal/prefab-internal.proto -I ../prefab-cloud-internal -I ../prefab-cloud/ --go_opt=Mprefab-internal.proto=prefab.cloud

mkdir -p internal-proto

protoc --go_out=. \
  --go_out=proto --go_opt=paths=source_relative \
  ./prefab-internal.slim.proto \
  -I ./ \
  -I ../prefab-cloud-internal \
  -I ../prefab-cloud/ \
  --go_opt=Mprefab-internal.slim.proto=prefab.internal.slim/

mv prefab-internal.slim.pb.go internal-proto/prefab-internal.slim.pb.go
rm proto/prefab-internal.slim.pb.go
