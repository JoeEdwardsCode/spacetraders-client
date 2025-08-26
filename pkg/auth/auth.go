package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/schema"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/transport"
	"strings"
	"sync"
	"time"
)

// AuthManager handles authentication and token management
type AuthManager struct {
	httpClient *transport.HTTPClient
	token      string
	agent      *schema.Agent
	mutex      sync.RWMutex
}

// Config represents authentication configuration
type Config struct {
	HTTPClient *transport.HTTPClient
	Token      string // Optional: pre-existing token
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *Config) *AuthManager {
	if config == nil {
		config = &Config{}
	}

	if config.HTTPClient == nil {
		config.HTTPClient = transport.NewHTTPClient(transport.DefaultConfig())
	}

	return &AuthManager{
		httpClient: config.HTTPClient,
		token:      config.Token,
	}
}

// RegisterAgent registers a new agent and obtains an authentication token
func (a *AuthManager) RegisterAgent(ctx context.Context, callSign, faction string) (*schema.RegisterAgentResponse, error) {
	if callSign == "" {
		return nil, fmt.Errorf("call sign cannot be empty")
	}
	if faction == "" {
		return nil, fmt.Errorf("faction cannot be empty")
	}

	// Validate call sign format (3-14 characters, alphanumeric and underscores)
	if !isValidCallSign(callSign) {
		return nil, fmt.Errorf("invalid call sign format: must be 3-14 characters, alphanumeric and underscores only")
	}

	// Validate faction (basic validation)
	if !isValidFaction(faction) {
		return nil, fmt.Errorf("invalid faction: %s", faction)
	}

	req := &transport.Request{
		Method: "POST",
		Path:   "/register",
		Body: schema.RegisterAgentRequest{
			Symbol:  callSign,
			Faction: faction,
		},
	}

	resp, err := a.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("registration request failed: %w", err)
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registration response: %w", err)
	}

	// Parse the registration response data
	regRespData, err := parseRegistrationResponse(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registration data: %w", err)
	}

	// Store authentication data
	a.mutex.Lock()
	a.token = regRespData.Token
	a.agent = &regRespData.Agent
	a.httpClient.SetToken(a.token)
	a.mutex.Unlock()

	return regRespData, nil
}

// SetToken manually sets the authentication token
func (a *AuthManager) SetToken(token string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.token = token
	a.httpClient.SetToken(token)
}

// GetToken returns the current authentication token
func (a *AuthManager) GetToken() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.token
}

// IsAuthenticated returns true if we have a valid token
func (a *AuthManager) IsAuthenticated() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	return a.token != "" && !a.isTokenExpired()
}

// GetAgent returns the current agent information
func (a *AuthManager) GetAgent(ctx context.Context) (*schema.Agent, error) {
	// If we have cached agent data and it's recent, return it
	a.mutex.RLock()
	if a.agent != nil {
		// Return cached agent (in a real implementation, we might check if it's stale)
		agent := *a.agent
		a.mutex.RUnlock()
		return &agent, nil
	}
	a.mutex.RUnlock()

	// Fetch agent data from API
	req := &transport.Request{
		Method: "GET",
		Path:   "/my/agent",
	}

	resp, err := a.httpClient.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent response: %w", err)
	}

	agent, err := parseAgentData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent data: %w", err)
	}

	// Cache the agent data
	a.mutex.Lock()
	a.agent = agent
	a.mutex.Unlock()

	return agent, nil
}

// RefreshAgent refreshes the cached agent data
func (a *AuthManager) RefreshAgent(ctx context.Context) (*schema.Agent, error) {
	a.mutex.Lock()
	a.agent = nil // Clear cache
	a.mutex.Unlock()

	return a.GetAgent(ctx)
}

// ValidateToken validates the current token by making an API call
func (a *AuthManager) ValidateToken(ctx context.Context) error {
	if !a.IsAuthenticated() {
		return fmt.Errorf("no authentication token available")
	}

	_, err := a.GetAgent(ctx)
	if err != nil {
		// If it's an auth error, clear the token
		if transport.IsAuthError(err) {
			a.mutex.Lock()
			a.token = ""
			a.agent = nil
			a.httpClient.SetToken("")
			a.mutex.Unlock()
		}
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}

// ClearAuth clears all authentication data
func (a *AuthManager) ClearAuth() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.token = ""
	a.agent = nil
	a.httpClient.SetToken("")
}

// GetAuthHeader returns the authorization header value
func (a *AuthManager) GetAuthHeader() string {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.token == "" {
		return ""
	}

	return "Bearer " + a.token
}

// Helper functions

// isValidCallSign validates call sign format
func isValidCallSign(callSign string) bool {
	if len(callSign) < 3 || len(callSign) > 14 {
		return false
	}

	for _, r := range callSign {
		if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}

// isValidFaction validates faction names (basic validation)
func isValidFaction(faction string) bool {
	// Known factions - in a real implementation, this might come from the API
	validFactions := []string{
		"COSMIC", "VOID", "GALACTIC", "QUANTUM", "DOMINION",
		"ASTRO", "CORSAIRS", "OBSIDIAN", "AEGIS", "UNITED",
		"SOLITARY", "COBALT", "OMEGA", "ECHO", "LORDS",
		"CULT", "ANCIENTS", "SHADOW", "ETHERIC",
	}

	for _, valid := range validFactions {
		if strings.EqualFold(faction, valid) {
			return true
		}
	}

	return false
}

// isTokenExpired checks if the JWT token is expired
func (a *AuthManager) isTokenExpired() bool {
	// In a real implementation, we would parse the JWT and check the expiration
	// For now, we'll assume tokens don't expire during a session
	return false
}

// parseRegistrationResponse parses the registration response data
func parseRegistrationResponse(data interface{}) (*schema.RegisterAgentResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var regResp schema.RegisterAgentResponse
	if err := json.Unmarshal(jsonData, &regResp); err != nil {
		return nil, err
	}

	return &regResp, nil
}

// parseAgentData parses agent data from API response
func parseAgentData(data interface{}) (*schema.Agent, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var agent schema.Agent
	if err := json.Unmarshal(jsonData, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

// TokenInfo represents information about the current token
type TokenInfo struct {
	HasToken    bool          `json:"has_token"`
	IsValid     bool          `json:"is_valid"`
	Agent       *schema.Agent `json:"agent,omitempty"`
	LastChecked time.Time     `json:"last_checked"`
}

// GetTokenInfo returns information about the current authentication state
func (a *AuthManager) GetTokenInfo(ctx context.Context) *TokenInfo {
	info := &TokenInfo{
		HasToken:    a.GetToken() != "",
		LastChecked: time.Now(),
	}

	if info.HasToken {
		agent, err := a.GetAgent(ctx)
		if err == nil {
			info.IsValid = true
			info.Agent = agent
		}
	}

	return info
}
