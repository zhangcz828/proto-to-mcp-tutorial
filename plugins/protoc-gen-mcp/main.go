package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	mcpannotations "proto-to-mcp-tutorial/generated/go/mcp/protobuf"

	httpannotations "google.golang.org/genproto/googleapis/api/annotations"
)

func main() {
	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		// Extract MCP methods from the proto files
		mcpMethods := extractMCPMethods(gen)

		if len(mcpMethods) == 0 {
			return nil // No MCP methods found
		}

		// Generate main server file
		generateMCPServer(gen, mcpMethods)

		return nil
	})
}

type MCPMethod struct {
	Service     *protogen.Service
	Method      *protogen.Method
	ToolName    string
	Description string
	HTTPInfo    *HTTPInfo
	Input       *protogen.Message
	Output      *protogen.Message
	Parameters  []*MCPParameter
}

type MCPParameter struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type HTTPInfo struct {
	Method string
	Path   string
	Body   string
}

func extractMCPMethods(gen *protogen.Plugin) []*MCPMethod {
	var mcpMethods []*MCPMethod

	for _, file := range gen.Files {
		if !file.Generate {
			continue
		}

		for _, service := range file.Services {
			for _, method := range service.Methods {
				if hasMCPToolAnnotation(method) {
					mcpMethod := &MCPMethod{
						Service:     service,
						Method:      method,
						ToolName:    generateToolName(method),
						Description: extractDescription(method),
						HTTPInfo:    extractHTTPInfo(method),
						Input:       method.Input,
						Output:      method.Output,
						Parameters:  extractParameters(method.Input),
					}
					mcpMethods = append(mcpMethods, mcpMethod)
				}
			}
		}
	}

	return mcpMethods
}

func hasMCPToolAnnotation(method *protogen.Method) bool {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	if options == nil {
		return false
	}

	// Check if method has mcp.v1.tool annotation
	if !proto.HasExtension(options, mcpannotations.E_Tool) {
		return false
	}

	// Get annotation value and check if enabled
	toolOptions := proto.GetExtension(options, mcpannotations.E_Tool).(*mcpannotations.MCPToolOptions)
	return toolOptions != nil && toolOptions.Enabled
}

func extractParameters(inputType *protogen.Message) []*MCPParameter {
	var parameters []*MCPParameter

	if inputType != nil {
		for _, field := range inputType.Fields {
			param := &MCPParameter{
				Name:        string(field.Desc.Name()),
				Type:        getFieldType(field),
				Required:    isFieldRequired(field),
				Description: extractFieldDescription(field),
			}
			parameters = append(parameters, param)
		}
	}

	return parameters
}

func isFieldRequired(field *protogen.Field) bool {
	// In proto3, technically all fields are optional, but for business logic:
	// If the field not marked as optional keyword, consider it required
	if field.Desc.HasOptionalKeyword() {
		return false
	}

	return true
}

func getFieldType(field *protogen.Field) string {
	// Check if the field is repeated (array/list)
	if field.Desc.Cardinality() == protoreflect.Repeated {
		return "list"
	}

	switch field.Desc.Kind() {
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return "integer"
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.MessageKind:
		return "object"
	case protoreflect.EnumKind:
		return "string"
	default:
		return "string"
	}
}

func extractFieldDescription(field *protogen.Field) string {
	if field.Comments.Leading != "" {
		return strings.TrimSpace(string(field.Comments.Leading))
	}
	return ""
}

func generateToolName(method *protogen.Method) string {
	methodName := string(method.Desc.Name())

	// Convert to snake_case
	methodName = camelToSnake(methodName)

	return methodName
}

func camelToSnake(str string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

func extractDescription(method *protogen.Method) string {
	if method.Comments.Leading != "" {
		return strings.TrimSpace(string(method.Comments.Leading))
	}
	return fmt.Sprintf("Execute %s RPC method", method.Desc.Name())
}

func extractHTTPInfo(method *protogen.Method) *HTTPInfo {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	if options == nil {
		return nil
	}

	if !proto.HasExtension(options, httpannotations.E_Http) {
		return nil
	}

	httpRule := proto.GetExtension(options, httpannotations.E_Http).(*httpannotations.HttpRule)
	if httpRule == nil {
		return nil
	}

	info := &HTTPInfo{}

	switch pattern := httpRule.Pattern.(type) {
	case *httpannotations.HttpRule_Get:
		info.Method = "GET"
		info.Path = pattern.Get
	case *httpannotations.HttpRule_Post:
		info.Method = "POST"
		info.Path = pattern.Post
		info.Body = httpRule.Body
	case *httpannotations.HttpRule_Put:
		info.Method = "PUT"
		info.Path = pattern.Put
		info.Body = httpRule.Body
	case *httpannotations.HttpRule_Delete:
		info.Method = "DELETE"
		info.Path = pattern.Delete
	case *httpannotations.HttpRule_Patch:
		info.Method = "PATCH"
		info.Path = pattern.Patch
		info.Body = httpRule.Body
	}

	return info
}

func generateMCPServer(gen *protogen.Plugin, mcpMethods []*MCPMethod) {
	outputFile := gen.NewGeneratedFile("mcp_server.py", ".")

	funcMap := template.FuncMap{
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"printf": fmt.Sprintf,
		"indent": func(text string, spaces int) string {
			if text == "" {
				return text
			}
			lines := strings.Split(text, "\n")
			indentStr := strings.Repeat(" ", spaces)
			var result []string
			for i, line := range lines {
				if strings.TrimSpace(line) != "" {
					result = append(result, indentStr+line)
				} else if i < len(lines)-1 { // Keep empty lines except the last one
					result = append(result, "")
				}
			}
			return strings.Join(result, "\n")
		},
	}

	tmpl := template.Must(template.New("mcp_server").Funcs(funcMap).Parse(mcpServerTemplate))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, mcpMethods); err != nil {
		return
	}

	outputFile.P(buf.String())
}

const mcpServerTemplate = `#!/usr/bin/env python3
"""
MCP Server for UPM API - Auto-generated from Protocol Buffers

This server provides access to project management operations
through the Model Context Protocol.
"""

import os
import sys
from typing import Any
import json

import httpx
from mcp.server.fastmcp import FastMCP

API_BASE = 'http://localhost:8080'
VERIFY_SSL = False

# Initialize FastMCP
mcp = FastMCP('Bookstore Server')

async def make_api_request(url: str, method: str = "GET", payload: dict = None) -> dict[str, Any] | None:
    """Make a HTTP request to the specified URL."""

    headers = {
        "Content-Type": "application/json",
    }
    # Use SSL verification based on environment variable
    async with httpx.AsyncClient(verify=VERIFY_SSL) as client:
        try:
            if method.upper() == "GET":
                response = await client.get(url, headers=headers, timeout=30.0)
            elif method.upper() == "PUT":
                response = await client.put(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "POST":
                response = await client.post(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "DELETE":
                response = await client.delete(url, headers=headers, timeout=30.0)
            elif method.upper() == "PATCH":
                response = await client.patch(url, headers=headers, json=payload, timeout=30.0)
            else:
                return {"error": f"Unsupported HTTP method: {method}"}
            
            response.raise_for_status()
            
            # Handle DELETE responses that might be empty
            if method.upper() == "DELETE":
                if response.status_code == 200 or response.status_code == 204:
                    return {"success": True, "message": "Resource deleted successfully"}
                
            # Try to parse JSON, return empty dict if no content
            try:
                return response.json()
            except:
                return {"success": True}
                
        except httpx.HTTPStatusError as e:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}
        except Exception as e:
            return {"error": str(e)}

# MCP Tools

{{range .}}
@mcp.tool()
async def {{.ToolName}}({{range $i, $param := .Parameters}}{{if $i}}, {{end}}{{$param.Name}}: {{if eq $param.Type "string"}}str{{else if eq $param.Type "integer"}}int{{else if eq $param.Type "boolean"}}bool{{else if eq $param.Type "list"}}list{{else}}dict{{end}}{{if not $param.Required}} = None{{end}}{{end}}) -> str:
    """{{.Description}}
    {{if .HTTPInfo}}
    HTTP: {{.HTTPInfo.Method}} {{.HTTPInfo.Path}}{{end}}
    
    Parameters:{{range .Parameters}}
    - {{.Name}} ({{.Type}}{{if not .Required}}, optional{{end}}): {{if contains .Description "\n"}}{{indent .Description 6}}{{else}}{{.Description}}{{end}}{{end}}
    
    Returns:
    - str: JSON formatted response from the API containing the result or error information
    """
    try:
        {{if .HTTPInfo}}
        # Construct the URL
        url = f"{API_BASE}{{.HTTPInfo.Path}}"
        {{$httpInfo := .HTTPInfo}}{{range .Parameters}}{{if and .Required (contains $httpInfo.Path (printf "{%s}" .Name))}}
        url = url.replace("{" + "{{.Name}}" + "}", str({{.Name}})){{end}}{{end}}
        
        # Prepare payload for non-GET requests
        payload = {}
        {{range .Parameters}}{{if and (ne $httpInfo.Method "GET") (ne $httpInfo.Method "DELETE") (not (contains $httpInfo.Path (printf "{%s}" .Name)))}}
        {{if .Required}}payload["{{.Name}}"] = {{.Name}}{{else}}if {{.Name}} is not None:
            payload["{{.Name}}"] = {{.Name}}{{end}}{{end}}{{end}}
        
        # Make the API request
        result = await make_api_request(url, "{{.HTTPInfo.Method}}", payload if payload else None)
        {{else}}
        result = {"error": "No HTTP endpoint defined for this method"}{{end}}
        
        # Return formatted JSON response
        import json
        return json.dumps(result, indent=2)
        
    except Exception as e:
        # Handle any errors that occur during execution
        import json
        error_result = {
            "error": f"Tool execution failed: {str(e)}",
            "tool_name": "{{.ToolName}}",
            "error_type": type(e).__name__
        }
        return json.dumps(error_result, indent=2)

{{end}}
if __name__ == '__main__':
    # Run the MCP server
    mcp.run()
`
