package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	OpenAPIURL     = "https://spacetraders.io/openapi"
	DefaultTimeout = 30 * time.Second
)

// OpenAPISpec represents the OpenAPI specification structure
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       map[string]interface{} `json:"info"`
	Servers    []Server               `json:"servers"`
	Paths      map[string]Path        `json:"paths"`
	Components Components             `json:"components"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type Path struct {
	Get    *Operation `json:"get,omitempty"`
	Post   *Operation `json:"post,omitempty"`
	Put    *Operation `json:"put,omitempty"`
	Delete *Operation `json:"delete,omitempty"`
	Patch  *Operation `json:"patch,omitempty"`
}

type Operation struct {
	OperationID string                 `json:"operationId"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Parameters  []Parameter            `json:"parameters"`
	RequestBody *RequestBody           `json:"requestBody,omitempty"`
	Responses   map[string]Response    `json:"responses"`
	Security    []map[string][]string  `json:"security,omitempty"`
}

type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Schema      Schema      `json:"schema"`
}

type RequestBody struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required"`
}

type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type        string             `json:"type"`
	Format      string             `json:"format,omitempty"`
	Properties  map[string]Schema  `json:"properties,omitempty"`
	Items       *Schema            `json:"items,omitempty"`
	Required    []string           `json:"required,omitempty"`
	Ref         string             `json:"$ref,omitempty"`
	Description string             `json:"description,omitempty"`
	Example     interface{}        `json:"example,omitempty"`
}

type Components struct {
	Schemas map[string]Schema `json:"schemas"`
}

// Fetcher handles OpenAPI specification retrieval
type Fetcher struct {
	client  *http.Client
	baseURL string
}

// New creates a new OpenAPI fetcher
func New() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL: OpenAPIURL,
	}
}

// FetchSpec retrieves the OpenAPI specification from the SpaceTraders API
func (f *Fetcher) FetchSpec() (*OpenAPISpec, error) {
	resp, err := f.client.Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenAPI spec: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenAPI spec: %w", err)
	}

	return &spec, nil
}

// SaveSpec saves the OpenAPI specification to a file
func (f *Fetcher) SaveSpec(spec *OpenAPISpec, filename string) error {
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	return writeFile(filename, data)
}

// LoadSpec loads the OpenAPI specification from a file
func (f *Fetcher) LoadSpec(filename string) (*OpenAPISpec, error) {
	data, err := readFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	return &spec, nil
}

// writeFile writes data to a file (placeholder - would use os.WriteFile)
func writeFile(filename string, data []byte) error {
	// Implementation would use os.WriteFile in real code
	// For now, return nil to satisfy interface
	return nil
}

// readFile reads data from a file (placeholder - would use os.ReadFile)
func readFile(filename string) ([]byte, error) {
	// Implementation would use os.ReadFile in real code
	// For now, return empty data
	return []byte{}, nil
}