# Introduce
This is a local search lib, base on beleve.
- include doc add, search, agg, suggest
- search in local, sync doc for multi node

# Service para
```golang
type ServicePara struct {
    DataPath string
    RpcPort int //if setup, run as rpc service
    DictFile string
    AddDocQueueMode bool //add doc with queue mode
}
```

# How to use?
Please see client.go in the **example** sub dir.

# proto generate
protoc --go_out=plugins=grpc:. *.proto

# Install proto3

visit https://github.com/google/protobuf/releases 
 
./configure;make;make install

go get github.com/golang/protobuf/protoc-gen-go 

cd github.com/golang/protobuf/protoc-gen-go 

go build 

go install or `cp -f protoc-gen-go /usr/local/go/bin`


# tips
- Do not open same search index in multi processes, this will cause file locked.

# testing
go test -v -run="QueryDoc"
go test -bench="QueryDoc"
go test -bench="QueryDoc" -benchmem -benchtime=10s