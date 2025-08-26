package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/schema"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/transport"
	"strconv"
)

// EndpointManager handles all API endpoint operations
type EndpointManager struct {
	httpClient *transport.HTTPClient
}

// NewEndpointManager creates a new endpoint manager
func NewEndpointManager(httpClient *transport.HTTPClient) *EndpointManager {
	return &EndpointManager{
		httpClient: httpClient,
	}
}

// Ship Operations

// GetFleet retrieves all ships owned by the agent
func (e *EndpointManager) GetFleet(ctx context.Context, opts *schema.PaginationOptions) ([]schema.Ship, error) {
	req := &transport.Request{
		Method:      "GET",
		Path:        "/my/ships",
		QueryParams: buildPaginationParams(opts),
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal fleet response: %w", err)
	}

	ships, err := parseShipsData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ships data: %w", err)
	}

	return ships, nil
}

// GetShip retrieves information about a specific ship
func (e *EndpointManager) GetShip(ctx context.Context, shipSymbol string) (*schema.Ship, error) {
	req := &transport.Request{
		Method: "GET",
		Path:   "/my/ships/" + shipSymbol,
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ship response: %w", err)
	}

	ship, err := parseShipData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ship data: %w", err)
	}

	return ship, nil
}

// OrbitShip puts a ship into orbit
func (e *EndpointManager) OrbitShip(ctx context.Context, shipSymbol string) (*schema.Ship, error) {
	req := &transport.Request{
		Method: "POST",
		Path:   "/my/ships/" + shipSymbol + "/orbit",
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal orbit response: %w", err)
	}

	// Extract nav data from response
	navData, err := parseNavData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nav data: %w", err)
	}

	// Return ship with updated nav (simplified for this implementation)
	ship := &schema.Ship{
		Symbol: shipSymbol,
		Nav:    *navData,
	}

	return ship, nil
}

// DockShip docks a ship at the current waypoint
func (e *EndpointManager) DockShip(ctx context.Context, shipSymbol string) (*schema.Ship, error) {
	req := &transport.Request{
		Method: "POST",
		Path:   "/my/ships/" + shipSymbol + "/dock",
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dock response: %w", err)
	}

	navData, err := parseNavData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nav data: %w", err)
	}

	ship := &schema.Ship{
		Symbol: shipSymbol,
		Nav:    *navData,
	}

	return ship, nil
}

// RefuelShip refuels a ship at the current waypoint
func (e *EndpointManager) RefuelShip(ctx context.Context, shipSymbol string) (*schema.Transaction, error) {
	req := &transport.Request{
		Method: "POST",
		Path:   "/my/ships/" + shipSymbol + "/refuel",
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refuel response: %w", err)
	}

	transaction, err := parseTransactionData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction data: %w", err)
	}

	return transaction, nil
}

// NavigateShip navigates a ship to a waypoint
func (e *EndpointManager) NavigateShip(ctx context.Context, shipSymbol, waypointSymbol string) (*schema.Navigation, error) {
	req := &transport.Request{
		Method: "POST",
		Path:   "/my/ships/" + shipSymbol + "/navigate",
		Body: schema.NavigateShipRequest{
			WaypointSymbol: waypointSymbol,
		},
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal navigate response: %w", err)
	}

	nav, err := parseNavData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse navigation data: %w", err)
	}

	return nav, nil
}

// GetShipNav gets the navigation information for a ship
func (e *EndpointManager) GetShipNav(ctx context.Context, shipSymbol string) (*schema.Navigation, error) {
	req := &transport.Request{
		Method: "GET",
		Path:   "/my/ships/" + shipSymbol + "/nav",
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nav response: %w", err)
	}

	nav, err := parseNavData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nav data: %w", err)
	}

	return nav, nil
}

// GetShipCargo gets the cargo information for a ship
func (e *EndpointManager) GetShipCargo(ctx context.Context, shipSymbol string) (*schema.Cargo, error) {
	req := &transport.Request{
		Method: "GET",
		Path:   "/my/ships/" + shipSymbol + "/cargo",
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cargo response: %w", err)
	}

	cargo, err := parseCargoData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cargo data: %w", err)
	}

	return cargo, nil
}

// Market Operations

// GetMarket retrieves market information for a waypoint
func (e *EndpointManager) GetMarket(ctx context.Context, systemSymbol, waypointSymbol string) (*schema.Market, error) {
	req := &transport.Request{
		Method: "GET",
		Path:   "/systems/" + systemSymbol + "/waypoints/" + waypointSymbol + "/market",
	}

	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal market response: %w", err)
	}

	market, err := parseMarketData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse market data: %w", err)
	}

	return market, nil
}

// PurchaseCargo purchases cargo from a market
func (e *EndpointManager) PurchaseCargo(ctx context.Context, shipSymbol string, req *schema.PurchaseCargoRequest) (*schema.Transaction, error) {
	httpReq := &transport.Request{
		Method: "POST",
		Path:   "/my/ships/" + shipSymbol + "/purchase",
		Body:   req,
	}

	resp, err := e.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal purchase response: %w", err)
	}

	transaction, err := parseTransactionData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction data: %w", err)
	}

	return transaction, nil
}

// SellCargo sells cargo to a market
func (e *EndpointManager) SellCargo(ctx context.Context, shipSymbol string, req *schema.SellCargoRequest) (*schema.Transaction, error) {
	httpReq := &transport.Request{
		Method: "POST",
		Path:   "/my/ships/" + shipSymbol + "/sell",
		Body:   req,
	}

	resp, err := e.httpClient.Do(ctx, httpReq)
	if err != nil {
		return nil, err
	}

	var apiResp schema.APIResponse
	if err := json.Unmarshal(resp.Body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sell response: %w", err)
	}

	transaction, err := parseTransactionData(apiResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction data: %w", err)
	}

	return transaction, nil
}

// Contract Operations (simplified implementations)

func (e *EndpointManager) GetContracts(ctx context.Context, opts *schema.PaginationOptions) ([]schema.Contract, error) {
	// Implementation similar to GetFleet but for contracts
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) GetContract(ctx context.Context, contractID string) (*schema.Contract, error) {
	// Implementation similar to GetShip but for contracts
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) AcceptContract(ctx context.Context, contractID string) (*schema.Contract, error) {
	// Implementation similar to OrbitShip but for contracts
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) DeliverContract(ctx context.Context, contractID, shipSymbol, tradeSymbol string, units int) (*schema.Contract, error) {
	// Implementation for contract delivery
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) FulfillContract(ctx context.Context, contractID string) (*schema.Contract, error) {
	// Implementation for contract fulfillment
	return nil, fmt.Errorf("not implemented")
}

// System Operations (simplified implementations)

func (e *EndpointManager) GetSystems(ctx context.Context, opts *schema.PaginationOptions) ([]schema.System, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) GetSystem(ctx context.Context, systemSymbol string) (*schema.System, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) GetWaypoints(ctx context.Context, systemSymbol string, opts *schema.PaginationOptions) ([]schema.Waypoint, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) GetWaypoint(ctx context.Context, systemSymbol, waypointSymbol string) (*schema.Waypoint, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) CreateSurvey(ctx context.Context, shipSymbol string) (*schema.Survey, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) ExtractResources(ctx context.Context, shipSymbol string, survey *schema.Survey) (*schema.Extraction, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) GetFactions(ctx context.Context, opts *schema.PaginationOptions) ([]schema.Faction, error) {
	return nil, fmt.Errorf("not implemented")
}

func (e *EndpointManager) GetFaction(ctx context.Context, factionSymbol string) (*schema.Faction, error) {
	return nil, fmt.Errorf("not implemented")
}

// Helper functions for parsing API responses

func buildPaginationParams(opts *schema.PaginationOptions) map[string]string {
	if opts == nil {
		return nil
	}

	params := make(map[string]string)
	if opts.Page != nil {
		params["page"] = strconv.Itoa(*opts.Page)
	}
	if opts.Limit != nil {
		params["limit"] = strconv.Itoa(*opts.Limit)
	}

	return params
}

func parseShipsData(data interface{}) ([]schema.Ship, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var ships []schema.Ship
	if err := json.Unmarshal(jsonData, &ships); err != nil {
		return nil, err
	}

	return ships, nil
}

func parseShipData(data interface{}) (*schema.Ship, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var ship schema.Ship
	if err := json.Unmarshal(jsonData, &ship); err != nil {
		return nil, err
	}

	return &ship, nil
}

func parseNavData(data interface{}) (*schema.Navigation, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var nav schema.Navigation
	if err := json.Unmarshal(jsonData, &nav); err != nil {
		return nil, err
	}

	return &nav, nil
}

func parseCargoData(data interface{}) (*schema.Cargo, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var cargo schema.Cargo
	if err := json.Unmarshal(jsonData, &cargo); err != nil {
		return nil, err
	}

	return &cargo, nil
}

func parseMarketData(data interface{}) (*schema.Market, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var market schema.Market
	if err := json.Unmarshal(jsonData, &market); err != nil {
		return nil, err
	}

	return &market, nil
}

func parseTransactionData(data interface{}) (*schema.Transaction, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var transaction schema.Transaction
	if err := json.Unmarshal(jsonData, &transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}
