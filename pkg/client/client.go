// Package client provides a comprehensive Go client for the SpaceTraders API.
//
// SpaceTraders is a space trading game where players create agents, manage fleets,
// trade goods, complete contracts, and explore the universe. This client provides
// a complete interface to interact with the SpaceTraders API.
//
// Basic usage:
//
//	import "github.com/JoeEdwardsCode/spacetraders-client/pkg/client"
//
//	// Create a new client
//	config := client.DefaultConfig()
//	client, err := client.New(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Register a new agent
//	ctx := context.Background()
//	resp, err := client.RegisterAgent(ctx, "MY_AGENT", "COSMIC")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get agent information
//	agent, err := client.GetAgent(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get fleet information
//	ships, err := client.GetFleet(ctx, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// The client provides comprehensive support for:
//   - Agent registration and management
//   - Fleet operations (ships, navigation, fuel)
//   - Market operations (buying, selling cargo)
//   - Contract management and fulfillment
//   - System and waypoint exploration
//   - Mining and survey operations
//   - Faction information
//   - Authentication and rate limiting
//
// Features:
//   - Automatic rate limiting following API guidelines
//   - Comprehensive error handling with typed errors
//   - Context support for timeouts and cancellation
//   - Mock server for testing and development
//   - Thread-safe operations with proper synchronization
//   - Configurable HTTP timeouts and retry logic
package client

import (
	"context"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/auth"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/endpoints"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/schema"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/transport"
	"time"
)

// SpaceTradersClient represents the main API client
type SpaceTradersClient struct {
	auth      *auth.AuthManager
	endpoints *endpoints.EndpointManager
	config    *Config
}

// Config represents client configuration
type Config struct {
	BaseURL   string
	Timeout   time.Duration
	UserAgent string
	Token     string // Optional: pre-existing token
}

// DefaultConfig returns a default client configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:   "https://api.spacetraders.io/v2",
		Timeout:   30 * time.Second,
		UserAgent: "SpaceTraders-Go-Client/1.0",
	}
}

// New creates a new SpaceTraders API client
func New(config *Config) (*SpaceTradersClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client
	httpConfig := transport.DefaultConfig()
	httpConfig.BaseURL = config.BaseURL
	httpConfig.Timeout = config.Timeout
	httpConfig.UserAgent = config.UserAgent
	httpClient := transport.NewHTTPClient(httpConfig)

	// Create auth manager
	authConfig := &auth.Config{
		HTTPClient: httpClient,
		Token:      config.Token,
	}
	authManager := auth.NewAuthManager(authConfig)

	// Create endpoint manager
	endpointManager := endpoints.NewEndpointManager(httpClient)

	return &SpaceTradersClient{
		auth:      authManager,
		endpoints: endpointManager,
		config:    config,
	}, nil
}

// Agent Operations

// RegisterAgent registers a new agent and obtains an authentication token
func (c *SpaceTradersClient) RegisterAgent(ctx context.Context, callSign, faction string) (*schema.RegisterAgentResponse, error) {
	return c.auth.RegisterAgent(ctx, callSign, faction)
}

// GetAgent retrieves the current agent information
func (c *SpaceTradersClient) GetAgent(ctx context.Context) (*schema.Agent, error) {
	return c.auth.GetAgent(ctx)
}

// SetToken manually sets the authentication token
func (c *SpaceTradersClient) SetToken(token string) {
	c.auth.SetToken(token)
}

// GetToken returns the current authentication token
func (c *SpaceTradersClient) GetToken() string {
	return c.auth.GetToken()
}

// IsAuthenticated returns true if the client has a valid authentication token
func (c *SpaceTradersClient) IsAuthenticated() bool {
	return c.auth.IsAuthenticated()
}

// Ship Operations

// GetFleet retrieves all ships owned by the agent
func (c *SpaceTradersClient) GetFleet(ctx context.Context, opts *schema.PaginationOptions) ([]schema.Ship, error) {
	return c.endpoints.GetFleet(ctx, opts)
}

// GetShip retrieves information about a specific ship
func (c *SpaceTradersClient) GetShip(ctx context.Context, shipSymbol string) (*schema.Ship, error) {
	return c.endpoints.GetShip(ctx, shipSymbol)
}

// OrbitShip puts a ship into orbit
func (c *SpaceTradersClient) OrbitShip(ctx context.Context, shipSymbol string) (*schema.Ship, error) {
	return c.endpoints.OrbitShip(ctx, shipSymbol)
}

// DockShip docks a ship at the current waypoint
func (c *SpaceTradersClient) DockShip(ctx context.Context, shipSymbol string) (*schema.Ship, error) {
	return c.endpoints.DockShip(ctx, shipSymbol)
}

// RefuelShip refuels a ship at the current waypoint
func (c *SpaceTradersClient) RefuelShip(ctx context.Context, shipSymbol string) (*schema.Transaction, error) {
	return c.endpoints.RefuelShip(ctx, shipSymbol)
}

// NavigateShip navigates a ship to a waypoint
func (c *SpaceTradersClient) NavigateShip(ctx context.Context, shipSymbol, waypointSymbol string) (*schema.Navigation, error) {
	return c.endpoints.NavigateShip(ctx, shipSymbol, waypointSymbol)
}

// GetShipNav gets the navigation information for a ship
func (c *SpaceTradersClient) GetShipNav(ctx context.Context, shipSymbol string) (*schema.Navigation, error) {
	return c.endpoints.GetShipNav(ctx, shipSymbol)
}

// GetShipCargo gets the cargo information for a ship
func (c *SpaceTradersClient) GetShipCargo(ctx context.Context, shipSymbol string) (*schema.Cargo, error) {
	return c.endpoints.GetShipCargo(ctx, shipSymbol)
}

// Market Operations

// GetMarket retrieves market information for a waypoint
func (c *SpaceTradersClient) GetMarket(ctx context.Context, systemSymbol, waypointSymbol string) (*schema.Market, error) {
	return c.endpoints.GetMarket(ctx, systemSymbol, waypointSymbol)
}

// PurchaseCargo purchases cargo from a market
func (c *SpaceTradersClient) PurchaseCargo(ctx context.Context, shipSymbol string, req *schema.PurchaseCargoRequest) (*schema.Transaction, error) {
	return c.endpoints.PurchaseCargo(ctx, shipSymbol, req)
}

// SellCargo sells cargo to a market
func (c *SpaceTradersClient) SellCargo(ctx context.Context, shipSymbol string, req *schema.SellCargoRequest) (*schema.Transaction, error) {
	return c.endpoints.SellCargo(ctx, shipSymbol, req)
}

// Contract Operations

// GetContracts retrieves all contracts available to the agent
func (c *SpaceTradersClient) GetContracts(ctx context.Context, opts *schema.PaginationOptions) ([]schema.Contract, error) {
	return c.endpoints.GetContracts(ctx, opts)
}

// GetContract retrieves information about a specific contract
func (c *SpaceTradersClient) GetContract(ctx context.Context, contractID string) (*schema.Contract, error) {
	return c.endpoints.GetContract(ctx, contractID)
}

// AcceptContract accepts a contract
func (c *SpaceTradersClient) AcceptContract(ctx context.Context, contractID string) (*schema.Contract, error) {
	return c.endpoints.AcceptContract(ctx, contractID)
}

// DeliverContract delivers cargo for a contract
func (c *SpaceTradersClient) DeliverContract(ctx context.Context, contractID, shipSymbol, tradeSymbol string, units int) (*schema.Contract, error) {
	return c.endpoints.DeliverContract(ctx, contractID, shipSymbol, tradeSymbol, units)
}

// FulfillContract fulfills a contract
func (c *SpaceTradersClient) FulfillContract(ctx context.Context, contractID string) (*schema.Contract, error) {
	return c.endpoints.FulfillContract(ctx, contractID)
}

// System & Exploration Operations

// GetSystems retrieves all systems
func (c *SpaceTradersClient) GetSystems(ctx context.Context, opts *schema.PaginationOptions) ([]schema.System, error) {
	return c.endpoints.GetSystems(ctx, opts)
}

// GetSystem retrieves information about a specific system
func (c *SpaceTradersClient) GetSystem(ctx context.Context, systemSymbol string) (*schema.System, error) {
	return c.endpoints.GetSystem(ctx, systemSymbol)
}

// GetWaypoints retrieves all waypoints in a system
func (c *SpaceTradersClient) GetWaypoints(ctx context.Context, systemSymbol string, opts *schema.PaginationOptions) ([]schema.Waypoint, error) {
	return c.endpoints.GetWaypoints(ctx, systemSymbol, opts)
}

// GetWaypoint retrieves information about a specific waypoint
func (c *SpaceTradersClient) GetWaypoint(ctx context.Context, systemSymbol, waypointSymbol string) (*schema.Waypoint, error) {
	return c.endpoints.GetWaypoint(ctx, systemSymbol, waypointSymbol)
}

// Mining & Survey Operations

// CreateSurvey creates a survey at the current waypoint
func (c *SpaceTradersClient) CreateSurvey(ctx context.Context, shipSymbol string) (*schema.Survey, error) {
	return c.endpoints.CreateSurvey(ctx, shipSymbol)
}

// ExtractResources extracts resources at the current waypoint
func (c *SpaceTradersClient) ExtractResources(ctx context.Context, shipSymbol string, survey *schema.Survey) (*schema.Extraction, error) {
	return c.endpoints.ExtractResources(ctx, shipSymbol, survey)
}

// Faction Operations

// GetFactions retrieves all factions
func (c *SpaceTradersClient) GetFactions(ctx context.Context, opts *schema.PaginationOptions) ([]schema.Faction, error) {
	return c.endpoints.GetFactions(ctx, opts)
}

// GetFaction retrieves information about a specific faction
func (c *SpaceTradersClient) GetFaction(ctx context.Context, factionSymbol string) (*schema.Faction, error) {
	return c.endpoints.GetFaction(ctx, factionSymbol)
}

// Utility Methods

// ValidateToken validates the current authentication token
func (c *SpaceTradersClient) ValidateToken(ctx context.Context) error {
	return c.auth.ValidateToken(ctx)
}

// GetTokenInfo returns information about the current authentication state
func (c *SpaceTradersClient) GetTokenInfo(ctx context.Context) *auth.TokenInfo {
	return c.auth.GetTokenInfo(ctx)
}

// GetRateLimiterState returns the current state of the rate limiter
func (c *SpaceTradersClient) GetRateLimiterState() interface{} {
	// This would return the actual rate limiter state
	// For now, return a placeholder
	return map[string]interface{}{
		"tokens_available": true,
		"next_refill":      time.Now().Add(time.Second),
	}
}

// Close closes the client and cleans up resources
func (c *SpaceTradersClient) Close() error {
	// In a real implementation, this might close HTTP connections, etc.
	c.auth.ClearAuth()
	return nil
}
