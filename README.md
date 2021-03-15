# Introduce

This is a local search lib, base on beleve.
- include doc add, search, agg, suggest
- search in local, sync doc for multi node

# How to use?

Please see client.go in the **example** sub dir.

# proto generate
protoc --go_out=plugins=grpc:. *.proto

# Install proto3
===
 https://github.com/google/protobuf/releases 
./configure;make;make install

go get github.com/golang/protobuf/protoc-gen-go 

cd github.com/golang/protobuf/protoc-gen-go 

go build 

go install or `cp -f protoc-gen-go /usr/local/go/bin`