# prefab-cloud-go
Go client for prefab




## proto compiling

update the git submodule in ./proto-source
`protoc --go_opt=paths=source_relative --go_out=. -I proto-source/  prefab.proto`