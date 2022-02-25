@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
@cd api
@.\protoc\bin\protoc.exe --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative ./mapassist.proto