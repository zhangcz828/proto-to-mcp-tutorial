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

		// Always generate file, even if no methods
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
				// Check if the method has MCP tool annotation, fallback to HTTP annotation
				if hasMCPToolAnnotation(method) || hasHTTPAnnotation(method) {
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

	// Debug: print number of methods found (removed for clean output)
	// fmt.Printf("DEBUG: Found %d MCP methods\n", len(mcpMethods))

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

func hasHTTPAnnotation(method *protogen.Method) bool {
	options := method.Desc.Options().(*descriptorpb.MethodOptions)
	if options == nil {
		return false
	}
	return proto.HasExtension(options, httpannotations.E_Http)
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
	// - Message type fields are usually required (like Book object)
	// - Primary key fields are usually required
	// - Path parameters are usually required

	fieldName := string(field.Desc.Name())

	// Message types are typically required
	if field.Desc.Kind() == protoreflect.MessageKind {
		return true
	}

	// ID fields are typically required
	if strings.Contains(strings.ToLower(fieldName), "id") {
		return true
	}

	// By default, consider fields optional
	return false
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
		"contains": strings.Contains,
		"printf":   fmt.Sprintf,
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("mcp_server").Funcs(funcMap).Parse(mcpServerTemplate))

	data := struct {
		ServiceName string
		Methods     []*MCPMethod
	}{}

	if len(mcpMethods) > 0 {
		data.ServiceName = string(mcpMethods[0].Service.Desc.Name())
	} else {
		data.ServiceName = "BookstoreService" // Default service name
	}
	data.Methods = mcpMethods

	if err := tmpl.Execute(&buf, data); err != nil {
		// Write error to output file for debugging
		outputFile.P("# Template execution error: " + err.Error())
		return
	}

	outputFile.P(buf.String())
}

const mcpServerTemplate = `#!/usr/bin/env python3
"""
MCP Server for {{.ServiceName}} - Auto-generated from Protocol Buffers
"""

from typing import Any, Dict, Optional
import httpx
from mcp.server.fastmcp import FastMCP

# Configuration
VERIFY_SSL: bool = False

# Initialize FastMCP
mcp = FastMCP("{{.ServiceName}} MCP Server")

async def make_api_request(url: str, method: str = "GET", payload: Optional[dict] = None) -> Dict[str, Any]:
    """Make an HTTP request to the service API."""
    headers = {"Content-Type": "application/json"}
    
    async with httpx.AsyncClient(verify=VERIFY_SSL) as client:
        try:
            if method.upper() == "GET":
                response = await client.get(url, headers=headers, timeout=30.0)
            elif method.upper() == "POST":
                response = await client.post(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "PUT":
                response = await client.put(url, headers=headers, json=payload, timeout=30.0)
            elif method.upper() == "DELETE":
                response = await client.delete(url, headers=headers, timeout=30.0)
            elif method.upper() == "PATCH":
                response = await client.patch(url, headers=headers, json=payload, timeout=30.0)
            else:
                return {"error": f"Unsupported HTTP method: {method}"}

            response.raise_for_status()
            
            try:
                return response.json()
            except Exception:
                return {"success": True}

        except httpx.HTTPStatusError as e:
            return {"error": f"HTTP {e.response.status_code}: {e.response.text}"}
        except Exception as e:
            return {"error": str(e)}

# MCP Tools
{{range $method := .Methods}}
@mcp.tool()
async def {{$method.ToolName}}({{range $i, $param := $method.Parameters}}{{if $i}}, {{end}}{{$param.Name}}: {{if $param.Required}}{{if eq $param.Type "string"}}str{{else if eq $param.Type "integer"}}int{{else if eq $param.Type "boolean"}}bool{{else if eq $param.Type "list"}}list{{else}}dict{{end}}{{else}}Optional[{{if eq $param.Type "string"}}str{{else if eq $param.Type "integer"}}int{{else if eq $param.Type "boolean"}}bool{{else if eq $param.Type "list"}}list{{else}}dict{{end}}] = None{{end}}{{end}}) -> str:
    """{{$method.Description}}

    {{if $method.HTTPInfo}}HTTP: {{$method.HTTPInfo.Method}} {{$method.HTTPInfo.Path}}{{end}}
    {{if $method.Parameters}}
    Parameters:{{range $param := $method.Parameters}}
    - {{$param.Name}} ({{$param.Type}}): {{if $param.Description}}{{$param.Description}}{{else}}{{$param.Name}} parameter{{end}}{{end}}{{end}}

    Returns:
    - str: JSON formatted response from the API containing the result or error information
    """
    try:
        {{if $method.HTTPInfo}}url = "http://localhost:8080{{$method.HTTPInfo.Path}}"
        
        # Replace path parameters
        {{range $param := $method.Parameters}}{{if and (ne $param.Name "") (contains $method.HTTPInfo.Path (printf "{%s}" $param.Name))}}{{if $param.Required}}url = url.replace("{{printf "{%s}" $param.Name}}", str({{$param.Name}})){{else}}if {{$param.Name}} is not None:
            url = url.replace("{{printf "{%s}" $param.Name}}", str({{$param.Name}})){{end}}
        {{end}}{{end}}
        
        # Prepare request body
        payload = None
        if "{{$method.HTTPInfo.Method}}" not in ["GET", "DELETE"]:
            body = {}
            {{range $param := $method.Parameters}}{{if not (contains $method.HTTPInfo.Path (printf "{%s}" $param.Name))}}{{if $param.Required}}body["{{$param.Name}}"] = {{$param.Name}}{{else}}if {{$param.Name}} is not None:
                body["{{$param.Name}}"] = {{$param.Name}}{{end}}
            {{end}}{{end}}if body:
                payload = body
        
        result = await make_api_request(url, "{{$method.HTTPInfo.Method}}", payload){{else}}result = {"error": "No HTTP endpoint defined"}{{end}}
        
        import json
        return json.dumps(result, indent=2)
        
    except Exception as e:
        import json
        return json.dumps({"error": f"Tool execution failed: {str(e)}"}, indent=2)

{{end}}
if __name__ == '__main__':
    mcp.run()
`
