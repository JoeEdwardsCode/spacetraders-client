package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JoeEdwardsCode/spacetraders-client/internal/ratelimit"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/schema"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.spacetraders.io/v2"
	DefaultTimeout = 30 * time.Second
	UserAgent      = "SpaceTraders-Go-Client/1.0"
)

// HTTPClient handles HTTP communication with the SpaceTraders API
type HTTPClient struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter *ratelimit.TokenBucket
	token       string
	userAgent   string
}

// Config represents HTTP client configuration
type Config struct {
	BaseURL     string
	Timeout     time.Duration
	UserAgent   string
	RateLimiter *ratelimit.TokenBucket
}

// DefaultConfig returns a default HTTP client configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:     DefaultBaseURL,
		Timeout:     DefaultTimeout,
		UserAgent:   UserAgent,
		RateLimiter: ratelimit.NewTokenBucket(),
	}
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config *Config) *HTTPClient {
	if config == nil {
		config = DefaultConfig()
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &HTTPClient{
		baseURL:     strings.TrimRight(config.BaseURL, "/"),
		httpClient:  httpClient,
		rateLimiter: config.RateLimiter,
		userAgent:   config.UserAgent,
	}
}

// SetToken sets the authentication token
func (c *HTTPClient) SetToken(token string) {
	c.token = token
}

// GetToken returns the current authentication token
func (c *HTTPClient) GetToken() string {
	return c.token
}

// Request represents an HTTP request
type Request struct {
	Method      string
	Path        string
	Body        interface{}
	QueryParams map[string]string
	Headers     map[string]string
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Do executes an HTTP request with rate limiting
func (c *HTTPClient) Do(ctx context.Context, req *Request) (*Response, error) {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter cancelled: %w", err)
	}

	// Build HTTP request
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	// Execute request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	response := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       body,
	}

	// Handle rate limit responses
	if httpResp.StatusCode == http.StatusTooManyRequests {
		return response, c.handleRateLimitResponse(httpResp)
	}

	// Handle other error status codes
	if httpResp.StatusCode >= 400 {
		return response, c.parseAPIError(body, httpResp.StatusCode)
	}

	return response, nil
}

// buildRequest constructs an HTTP request from a Request object
func (c *HTTPClient) buildRequest(ctx context.Context, req *Request) (*http.Request, error) {
	// Build URL
	requestURL := c.baseURL + req.Path

	// Add query parameters
	if len(req.QueryParams) > 0 {
		params := url.Values{}
		for k, v := range req.QueryParams {
			params.Add(k, v)
		}
		requestURL += "?" + params.Encode()
	}

	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("User-Agent", c.userAgent)
	if req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set authentication header
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Set custom headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

// handleRateLimitResponse processes 429 responses and extracts rate limit information
func (c *HTTPClient) handleRateLimitResponse(resp *http.Response) error {
	retryAfterHeader := resp.Header.Get("Retry-After")
	rateLimitType := resp.Header.Get("x-ratelimit-type")

	var retryAfter time.Duration
	if retryAfterHeader != "" {
		if seconds, err := strconv.Atoi(retryAfterHeader); err == nil {
			retryAfter = time.Duration(seconds) * time.Second
		}
	}

	return &RateLimitError{
		Type:       rateLimitType,
		RetryAfter: retryAfter,
		Limit:      parseIntHeader(resp.Header.Get("x-ratelimit-limit")),
		Remaining:  parseIntHeader(resp.Header.Get("x-ratelimit-remaining")),
		Reset:      parseTimeHeader(resp.Header.Get("x-ratelimit-reset")),
	}
}

// parseAPIError parses API error responses
func (c *HTTPClient) parseAPIError(body []byte, statusCode int) error {
	var apiError schema.APIError
	if err := json.Unmarshal(body, &apiError); err != nil {
		// If we can't parse the error, return a generic one
		return &APIError{
			StatusCode: statusCode,
			Message:    string(body),
		}
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    apiError.Message,
		Code:       apiError.Code,
		Data:       apiError.Data,
	}
}

// GetRateLimiterState returns the current state of the rate limiter
func (c *HTTPClient) GetRateLimiterState() ratelimit.BucketState {
	return c.rateLimiter.GetState()
}

// ResetRateLimiter resets the rate limiter to full capacity
func (c *HTTPClient) ResetRateLimiter() {
	c.rateLimiter.Reset()
}

// Utility functions

func parseIntHeader(value string) int {
	if value == "" {
		return 0
	}
	i, _ := strconv.Atoi(value)
	return i
}

func parseTimeHeader(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	// Try parsing as Unix timestamp
	if timestamp, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Unix(timestamp, 0)
	}

	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}

	return time.Time{}
}

// Error types

// RateLimitError represents a rate limit error from the API
type RateLimitError struct {
	Type       string        `json:"type"`
	RetryAfter time.Duration `json:"retry_after"`
	Limit      int           `json:"limit"`
	Remaining  int           `json:"remaining"`
	Reset      time.Time     `json:"reset"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded (type: %s, retry after: %v)", e.Type, e.RetryAfter)
}

// IsRateLimitError returns true if the error is a rate limit error
func IsRateLimitError(err error) bool {
	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}

// APIError represents a general API error
type APIError struct {
	StatusCode int                    `json:"status_code"`
	Message    string                 `json:"message"`
	Code       int                    `json:"code"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status: %d, code: %d): %s", e.StatusCode, e.Code, e.Message)
}

// IsAPIError returns true if the error is an API error
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// IsAuthError returns true if the error is an authentication error
func IsAuthError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}
