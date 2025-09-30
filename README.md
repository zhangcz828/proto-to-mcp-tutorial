# Proto to MCP Tutorial

## Overview
This project shows how to build an MCP (Model Context Protocol) server from existing Protocol Buffers definitions. A custom `protoc` plugin automatically generates the MCP server so that a single `.proto` source produces:
* gRPC services
* OpenAPI specifications
* REST endpoints
* MCP tools

## Blog Series (4 Parts)
The project is explained step by step in an accompanying blog series:

| Part | Title                                                                                                                                                                 |
| ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1    | [Generating REST APIs from Protobuf Using gRPC Transcoding](https://www.enterprisedb.com/blog/building-mcp-servers-protobuf-part1-protobuf-rest-api)                  |
| 2    | [Automate MCP Server Creation with Protoc Plugins](https://www.enterprisedb.com/blog/building-mcp-servers-protobuf-part2-automate-mcp-server-creation-protoc-plugins) |
| 3    | [Enhance AI Interactions with Proto Comments](https://www.enterprisedb.com/blog/building-mcp-servers-protobuf-part3-enhance-ai-interactions-proto-comments)           |
| 4    | Insights from Running MCP Tools in Practice (To Be Done)                                                                                                              |
