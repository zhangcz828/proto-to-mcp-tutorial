export GOOGLEAPIS_DIR=./googleapis

# Generate gRPC service code
protoc -I${GOOGLEAPIS_DIR} --proto_path=proto --go_out=./generated/go  --go_opt paths=source_relative --go-grpc_out=./generated/go --go-grpc_opt paths=source_relative bookstore.proto

# Generate gRPC Gateway for REST endpoints  
protoc -I${GOOGLEAPIS_DIR} --proto_path=proto --grpc-gateway_out=./generated/go --grpc-gateway_opt paths=source_relative bookstore.proto

# Generate OpenAPI specifications
protoc -I${GOOGLEAPIS_DIR} -I./proto --openapi_out=./generated/openapi \
      --openapi_opt=fq_schema_naming=true \
      --openapi_opt=version="1.0.0" \
      --openapi_opt=title="Bookstore API" \
      bookstore.proto