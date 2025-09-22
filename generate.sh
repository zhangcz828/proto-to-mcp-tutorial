#!/bin/bash

export GOOGLEAPIS_DIR=./googleapis
export MCP_DIR=.

echo "Checking for required protoc plugins..."

# Check if protoc-gen-go is available
if ! command -v protoc-gen-go &> /dev/null; then
    echo "❌ protoc-gen-go not found. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    if [ $? -ne 0 ]; then
        echo "⚠️  Failed to install protoc-gen-go. Skipping Go generation."
        exit 1
    fi
fi

# Check if protoc-gen-go-grpc is available  
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "❌ protoc-gen-go-grpc not found. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    if [ $? -ne 0 ]; then
        echo "⚠️  Failed to install protoc-gen-go-grpc. Skipping gRPC generation."
        exit 1
    fi
fi

# Check if protoc-gen-grpc-gateway is available
if ! command -v protoc-gen-grpc-gateway &> /dev/null; then
    echo "❌ protoc-gen-grpc-gateway not found. Installing..."
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
    if [ $? -ne 0 ]; then
        echo "⚠️  Failed to install protoc-gen-grpc-gateway. Skipping Gateway generation."
        exit 1
    fi
fi


echo "🔧 Generating Go gRPC code..."

# Step 1: Generate Go code for MCP annotations
echo "🔧 Generating MCP annotations Go code..."
protoc -I${GOOGLEAPIS_DIR} -I${MCP_DIR} --go_out=./generated/go --go_opt=paths=source_relative mcp/protobuf/annotations.proto

# Step 2: Generate Go gRPC code
protoc -I${GOOGLEAPIS_DIR} -I${MCP_DIR} --proto_path=proto --go_out=./generated/go  --go_opt paths=source_relative --go-grpc_out=./generated/go --go-grpc_opt paths=source_relative bookstore.proto


# Generate gRPC Gateway for REST endpoints
echo "🔧 Generating gRPC Gateway..."
protoc -I${GOOGLEAPIS_DIR} -I${MCP_DIR} --proto_path=proto --grpc-gateway_out=./generated/go --grpc-gateway_opt paths=source_relative bookstore.proto

# Generate OpenAPI specifications
echo "🔧 Generating OpenAPI spec..."
protoc -I${GOOGLEAPIS_DIR} -I${MCP_DIR} -I./proto --openapi_out=./generated/openapi \
      --openapi_opt=fq_schema_naming=true \
      --openapi_opt=version="1.0.0" \
      --openapi_opt=title="Bookstore API" \
      bookstore.proto

# Generate MCP server (always available since we have our custom plugin)
echo "🔧 Generating MCP server..."
protoc -I${GOOGLEAPIS_DIR} -I${MCP_DIR} --proto_path=proto \
      --plugin=protoc-gen-mcp=./protoc-gen-mcp \
      --mcp_out=./generated/mcp \
      bookstore.proto

echo ""
echo "🎉 Generation complete!"
echo "📁 Check the ./generated directory for all generated files."