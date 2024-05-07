#!/bin/bash

#run from the root directory

protoc --proto_path=proto-source \
  --go_out=proto --go_opt=paths=source_relative \
  --go_opt=Mproto-source/prefab.proto=github.com/prefab-cloud/prefab-cloud-go/proto \
  proto-source/prefab.proto
