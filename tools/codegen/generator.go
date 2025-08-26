package codegen

import (
	"fmt"
	"github.com/JoeEdwardsCode/spacetraders-client/tools/fetcher"
	"strings"
)

// Generator handles Go code generation from OpenAPI specifications
type Generator struct {
	spec *fetcher.OpenAPISpec
}

// New creates a new code generator
func New(spec *fetcher.OpenAPISpec) *Generator {
	return &Generator{spec: spec}
}

// GenerateTypes generates Go struct types from OpenAPI schemas
func (g *Generator) GenerateTypes() (string, error) {
	if g.spec == nil {
		return "", fmt.Errorf("no OpenAPI spec provided")
	}

	var builder strings.Builder

	// Package header
	builder.WriteString("// Code generated from OpenAPI specification. DO NOT EDIT.\n\n")
	builder.WriteString("package schema\n\n")
	builder.WriteString("import (\n")
	builder.WriteString("\t\"time\"\n")
	builder.WriteString(")\n\n")

	// Generate struct types for each schema
	for name, schema := range g.spec.Components.Schemas {
		structCode := g.generateStruct(name, schema)
		builder.WriteString(structCode)
		builder.WriteString("\n\n")
	}

	return builder.String(), nil
}

// generateStruct generates a Go struct from an OpenAPI schema
func (g *Generator) generateStruct(name string, schema fetcher.Schema) string {
	var builder strings.Builder

	// Add documentation if available
	if schema.Description != "" {
		builder.WriteString(fmt.Sprintf("// %s %s\n", name, schema.Description))
	}

	builder.WriteString(fmt.Sprintf("type %s struct {\n", toGoTypeName(name)))

	// Generate fields
	for fieldName, fieldSchema := range schema.Properties {
		fieldType := g.mapToGoType(fieldSchema)
		jsonTag := fmt.Sprintf("`json:\"%s\"`", fieldName)

		// Check if field is required
		isRequired := contains(schema.Required, fieldName)
		if !isRequired && fieldType != "string" && fieldType != "bool" {
			fieldType = "*" + fieldType // Make non-required fields pointers
		}

		builder.WriteString(fmt.Sprintf("\t%s %s %s\n",
			toGoFieldName(fieldName), fieldType, jsonTag))
	}

	builder.WriteString("}")

	return builder.String()
}

// mapToGoType maps OpenAPI types to Go types
func (g *Generator) mapToGoType(schema fetcher.Schema) string {
	switch schema.Type {
	case "string":
		if schema.Format == "date-time" {
			return "time.Time"
		}
		return "string"
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if schema.Format == "float" {
			return "float32"
		}
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		if schema.Items != nil {
			itemType := g.mapToGoType(*schema.Items)
			return "[]" + itemType
		}
		return "[]interface{}"
	case "object":
		if len(schema.Properties) > 0 {
			// For inline objects, could generate anonymous struct
			return "map[string]interface{}"
		}
		return "map[string]interface{}"
	}

	// Handle $ref references
	if schema.Ref != "" {
		return toGoTypeName(extractRefName(schema.Ref))
	}

	return "interface{}"
}

// GenerateEndpoints generates Go methods for API endpoints
func (g *Generator) GenerateEndpoints() (string, error) {
	var builder strings.Builder

	builder.WriteString("// Code generated from OpenAPI specification. DO NOT EDIT.\n\n")
	builder.WriteString("package endpoints\n\n")
	builder.WriteString("import (\n")
	builder.WriteString("\t\"context\"\n")
	builder.WriteString("\t\"github.com/JoeEdwardsCode/spacetraders-client/pkg/schema\"\n")
	builder.WriteString(")\n\n")

	// Generate interface
	builder.WriteString("// SpaceTradersAPI defines all API operations\n")
	builder.WriteString("type SpaceTradersAPI interface {\n")

	for path, pathItem := range g.spec.Paths {
		if pathItem.Get != nil {
			method := g.generateMethodSignature(path, "GET", pathItem.Get)
			builder.WriteString("\t" + method + "\n")
		}
		if pathItem.Post != nil {
			method := g.generateMethodSignature(path, "POST", pathItem.Post)
			builder.WriteString("\t" + method + "\n")
		}
		if pathItem.Put != nil {
			method := g.generateMethodSignature(path, "PUT", pathItem.Put)
			builder.WriteString("\t" + method + "\n")
		}
		if pathItem.Delete != nil {
			method := g.generateMethodSignature(path, "DELETE", pathItem.Delete)
			builder.WriteString("\t" + method + "\n")
		}
		if pathItem.Patch != nil {
			method := g.generateMethodSignature(path, "PATCH", pathItem.Patch)
			builder.WriteString("\t" + method + "\n")
		}
	}

	builder.WriteString("}\n")

	return builder.String(), nil
}

// generateMethodSignature generates a Go method signature from an OpenAPI operation
func (g *Generator) generateMethodSignature(path, method string, op *fetcher.Operation) string {
	methodName := toGoMethodName(op.OperationID)
	if methodName == "" {
		methodName = generateMethodName(method, path)
	}

	// Build parameters
	var params []string
	params = append(params, "ctx context.Context")

	// Add path parameters
	for _, param := range op.Parameters {
		if param.In == "path" {
			goType := mapParamToGoType(param.Schema)
			params = append(params, fmt.Sprintf("%s %s",
				toGoParamName(param.Name), goType))
		}
	}

	// Add query parameters as options struct
	hasQueryParams := false
	for _, param := range op.Parameters {
		if param.In == "query" {
			hasQueryParams = true
			break
		}
	}
	if hasQueryParams {
		params = append(params, "opts *QueryOptions")
	}

	// Add request body if present
	if op.RequestBody != nil {
		params = append(params, "body interface{}")
	}

	// Determine return type
	returnType := "error"
	for code, response := range op.Responses {
		if code == "200" || code == "201" {
			if len(response.Content) > 0 {
				returnType = "(*schema.Response, error)"
				break
			}
		}
	}

	paramStr := strings.Join(params, ", ")
	return fmt.Sprintf("%s(%s) %s", methodName, paramStr, returnType)
}

// Utility functions

func toGoTypeName(name string) string {
	return toPascalCase(name)
}

func toGoFieldName(name string) string {
	return toPascalCase(name)
}

func toGoMethodName(name string) string {
	return toPascalCase(name)
}

func toGoParamName(name string) string {
	return toCamelCase(name)
}

func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	var result strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])))
			if len(word) > 1 {
				result.WriteString(strings.ToLower(word[1:]))
			}
		}
	}
	return result.String()
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) > 0 {
		return strings.ToLower(string(pascal[0])) + pascal[1:]
	}
	return pascal
}

func extractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func mapParamToGoType(schema fetcher.Schema) string {
	switch schema.Type {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	default:
		return "string"
	}
}

func generateMethodName(httpMethod, path string) string {
	method := strings.ToLower(httpMethod)
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	var name strings.Builder
	name.WriteString(strings.Title(method))

	for _, part := range pathParts {
		if !strings.HasPrefix(part, "{") {
			name.WriteString(toPascalCase(part))
		}
	}

	return name.String()
}
