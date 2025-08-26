package mock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"spacetraders-client/internal/ratelimit"
	"spacetraders-client/pkg/schema"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MockServer simulates the SpaceTraders API with business logic
type MockServer struct {
	server      *httptest.Server
	rateLimiter *ratelimit.TokenBucket
	gameState   *GameState
	mutex       sync.RWMutex
}

// GameState represents the simulated game state
type GameState struct {
	Agents    map[string]*schema.Agent    `json:"agents"`
	Ships     map[string]*schema.Ship     `json:"ships"`
	Contracts map[string]*schema.Contract `json:"contracts"`
	Markets   map[string]*schema.Market   `json:"markets"`
	Systems   map[string]*schema.System   `json:"systems"`
	Waypoints map[string]*schema.Waypoint `json:"waypoints"`
	Tokens    map[string]string           `json:"tokens"` // token -> agent symbol
	
	// Business logic state
	FuelPrices    map[string]int `json:"fuel_prices"`    // waypoint -> price
	MarketPrices  map[string]map[string]int `json:"market_prices"` // waypoint -> good -> price
	TravelTimes   map[string]map[string]time.Duration `json:"travel_times"` // origin -> destination -> time
	LastUpdate    time.Time `json:"last_update"`
}

// NewMockServer creates a new mock SpaceTraders API server
func NewMockServer() *MockServer {
	gameState := &GameState{
		Agents:       make(map[string]*schema.Agent),
		Ships:        make(map[string]*schema.Ship),
		Contracts:    make(map[string]*schema.Contract),
		Markets:      make(map[string]*schema.Market),
		Systems:      make(map[string]*schema.System),
		Waypoints:    make(map[string]*schema.Waypoint),
		Tokens:       make(map[string]string),
		FuelPrices:   make(map[string]int),
		MarketPrices: make(map[string]map[string]int),
		TravelTimes:  make(map[string]map[string]time.Duration),
		LastUpdate:   time.Now(),
	}

	// Initialize with sample data
	gameState.initializeGameData()

	mock := &MockServer{
		rateLimiter: ratelimit.NewTokenBucket(),
		gameState:   gameState,
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mock.setupRoutes(mux)
	mock.server = httptest.NewServer(mux)

	return mock
}

// GetURL returns the mock server URL
func (m *MockServer) GetURL() string {
	return m.server.URL
}

// Close closes the mock server
func (m *MockServer) Close() {
	m.server.Close()
}

// SetRateLimitEnabled enables or disables rate limiting
func (m *MockServer) SetRateLimitEnabled(enabled bool) {
	if !enabled {
		m.rateLimiter = nil
	} else {
		m.rateLimiter = ratelimit.NewTokenBucket()
	}
}

// setupRoutes configures all the API routes
func (m *MockServer) setupRoutes(mux *http.ServeMux) {
	// Agent registration (no auth middleware)
	mux.HandleFunc("/register", m.withRateLimit(m.handleRegister))
	
	// Agent operations (with auth middleware)
	mux.HandleFunc("/my/agent", m.withMiddleware(m.handleGetAgent))
	
	// Ship operations (with auth middleware)
	mux.HandleFunc("/my/ships", m.withMiddleware(m.handleGetFleet))
	mux.HandleFunc("/my/ships/", m.withMiddleware(m.handleShipOperations))
	
	// Market operations (with auth middleware)
	mux.HandleFunc("/systems/", m.withMiddleware(m.handleSystemOperations))
	
	// Contract operations (with auth middleware)
	mux.HandleFunc("/my/contracts", m.withMiddleware(m.handleGetContracts))
	mux.HandleFunc("/my/contracts/", m.withMiddleware(m.handleContractOperations))
}

// Middleware for rate limiting and authentication
func (m *MockServer) withMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Rate limiting
		if m.rateLimiter != nil {
			if !m.rateLimiter.TryAllow() {
				m.writeRateLimitError(w)
				return
			}
		}

		// Authentication (except for registration)
		if r.URL.Path != "/register" {
			if !m.checkAuth(r) {
				m.writeAuthError(w)
				return
			}
		}

		// Call the actual handler
		handler(w, r)
	}
}

// Middleware for rate limiting only (for registration endpoint)
func (m *MockServer) withRateLimit(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Rate limiting
		if m.rateLimiter != nil {
			if !m.rateLimiter.TryAllow() {
				m.writeRateLimitError(w)
				return
			}
		}

		// Call the actual handler
		handler(w, r)
	}
}

// Agent registration handler
func (m *MockServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req schema.RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Symbol == "" || req.Faction == "" {
		m.writeError(w, http.StatusBadRequest, "Symbol and faction are required")
		return
	}

	// Check if agent already exists
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.gameState.Agents[req.Symbol]; exists {
		m.writeError(w, http.StatusConflict, "Agent already exists")
		return
	}

	// Create new agent
	agent := m.createAgent(req.Symbol, req.Faction)
	ship := m.createStartingShip(agent)
	contract := m.createStartingContract(agent)
	faction := m.getFaction(req.Faction)
	token := m.generateToken(agent.Symbol)

	// Store in game state
	m.gameState.Agents[agent.Symbol] = agent
	m.gameState.Ships[ship.Symbol] = ship
	m.gameState.Contracts[contract.ID] = contract
	m.gameState.Tokens[token] = agent.Symbol

	// Create response
	response := schema.RegisterAgentResponse{
		Agent:    *agent,
		Ship:     *ship,
		Contract: *contract,
		Faction:  *faction,
		Token:    token,
	}

	m.writeJSONResponse(w, http.StatusCreated, response)
}

// Get agent handler
func (m *MockServer) handleGetAgent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentSymbol := m.getAgentFromToken(r)
	if agentSymbol == "" {
		m.writeAuthError(w)
		return
	}

	m.mutex.RLock()
	agent, exists := m.gameState.Agents[agentSymbol]
	m.mutex.RUnlock()

	if !exists {
		m.writeError(w, http.StatusNotFound, "Agent not found")
		return
	}

	m.writeJSONResponse(w, http.StatusOK, *agent)
}

// Get fleet handler
func (m *MockServer) handleGetFleet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentSymbol := m.getAgentFromToken(r)
	if agentSymbol == "" {
		m.writeAuthError(w)
		return
	}

	m.mutex.RLock()
	var ships []schema.Ship
	for _, ship := range m.gameState.Ships {
		if strings.HasPrefix(ship.Symbol, agentSymbol+"-") {
			ships = append(ships, *ship)
		}
	}
	m.mutex.RUnlock()

	m.writeJSONResponse(w, http.StatusOK, ships)
}

// Ship operations handler
func (m *MockServer) handleShipOperations(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	shipSymbol := pathParts[2]
	
	if len(pathParts) == 3 {
		// GET /my/ships/{shipSymbol}
		if r.Method == http.MethodGet {
			m.handleGetShip(w, r, shipSymbol)
			return
		}
	}

	if len(pathParts) == 4 {
		operation := pathParts[3]
		switch operation {
		case "orbit":
			m.handleShipOrbit(w, r, shipSymbol)
		case "dock":
			m.handleShipDock(w, r, shipSymbol)
		case "navigate":
			m.handleShipNavigate(w, r, shipSymbol)
		case "refuel":
			m.handleShipRefuel(w, r, shipSymbol)
		case "purchase":
			m.handlePurchaseCargo(w, r, shipSymbol)
		case "sell":
			m.handleSellCargo(w, r, shipSymbol)
		default:
			http.Error(w, "Unknown operation", http.StatusNotFound)
		}
	}
}

// Business logic methods

func (m *MockServer) createAgent(symbol, faction string) *schema.Agent {
	return &schema.Agent{
		AccountID:       "mock-account-" + symbol,
		Symbol:          symbol,
		Headquarters:    faction + "-HQ",
		Credits:         150000, // Starting credits
		StartingFaction: faction,
		ShipCount:       1,
	}
}

func (m *MockServer) createStartingShip(agent *schema.Agent) *schema.Ship {
	return &schema.Ship{
		Symbol: agent.Symbol + "-1",
		Registration: schema.Registration{
			Name:          "Starting Ship",
			FactionSymbol: agent.StartingFaction,
			Role:          "COMMAND",
		},
		Nav: schema.Navigation{
			SystemSymbol:   "X1-TEST",
			WaypointSymbol: "X1-TEST-A1",
			Status:         "DOCKED",
			FlightMode:     "CRUISE",
		},
		Cargo: schema.Cargo{
			Capacity:  40,
			Units:     0,
			Inventory: []schema.CargoItem{},
		},
		Fuel: schema.Fuel{
			Current:  100,
			Capacity: 100,
		},
	}
}

func (m *MockServer) createStartingContract(agent *schema.Agent) *schema.Contract {
	return &schema.Contract{
		ID:            "contract-" + agent.Symbol + "-1",
		FactionSymbol: agent.StartingFaction,
		Type:          "PROCUREMENT",
		Terms: schema.ContractTerms{
			Deadline: time.Now().Add(7 * 24 * time.Hour),
			Payment: schema.ContractPayment{
				OnAccepted:  10000,
				OnFulfilled: 50000,
			},
			Deliver: []schema.ContractDeliverGood{
				{
					TradeSymbol:       "IRON",
					DestinationSymbol: "X1-TEST-A1",
					UnitsRequired:     100,
					UnitsFulfilled:    0,
				},
			},
		},
		Accepted:         false,
		Fulfilled:        false,
		Expiration:       time.Now().Add(24 * time.Hour),
		DeadlineToAccept: &[]time.Time{time.Now().Add(2 * time.Hour)}[0],
	}
}

func (m *MockServer) getFaction(symbol string) *schema.Faction {
	return &schema.Faction{
		Symbol:       symbol,
		Name:         symbol + " Faction",
		Description:  "A space-faring faction",
		Headquarters: symbol + "-HQ",
		Traits: []schema.FactionTrait{
			{
				Symbol:      "TRADERS",
				Name:        "Traders",
				Description: "Focused on trade and commerce",
			},
		},
		IsRecruiting: true,
	}
}

func (m *MockServer) generateToken(agentSymbol string) string {
	return "mock-token-" + agentSymbol + "-" + strconv.FormatInt(time.Now().Unix(), 10)
}

// Helper methods

func (m *MockServer) checkAuth(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	m.mutex.RLock()
	_, exists := m.gameState.Tokens[token]
	m.mutex.RUnlock()

	return exists
}

func (m *MockServer) getAgentFromToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	m.mutex.RLock()
	agentSymbol, exists := m.gameState.Tokens[token]
	m.mutex.RUnlock()

	if !exists {
		return ""
	}

	return agentSymbol
}

func (m *MockServer) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := schema.APIResponse{
		Data: data,
	}

	json.NewEncoder(w).Encode(response)
}

func (m *MockServer) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := schema.APIError{
		Message: message,
		Code:    statusCode,
	}

	json.NewEncoder(w).Encode(errorResp)
}

func (m *MockServer) writeRateLimitError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-ratelimit-type", "requests")
	w.Header().Set("x-ratelimit-limit", "30")
	w.Header().Set("x-ratelimit-remaining", "0")
	w.Header().Set("Retry-After", "1")
	w.WriteHeader(http.StatusTooManyRequests)

	errorResp := schema.APIError{
		Message: "Rate limit exceeded",
		Code:    http.StatusTooManyRequests,
	}

	json.NewEncoder(w).Encode(errorResp)
}

func (m *MockServer) writeAuthError(w http.ResponseWriter) {
	m.writeError(w, http.StatusUnauthorized, "Authentication required")
}

// Initialize game data with sample systems, markets, etc.
func (gs *GameState) initializeGameData() {
	// Add sample systems and waypoints
	system := &schema.System{
		Symbol:       "X1-TEST",
		SectorSymbol: "X1",
		Type:         "RED_STAR",
		X:            0,
		Y:            0,
		Waypoints:    []schema.Waypoint{},
	}
	gs.Systems[system.Symbol] = system

	// Add sample waypoints
	waypoint := &schema.Waypoint{
		Symbol:       "X1-TEST-A1",
		Type:         "PLANET",
		SystemSymbol: "X1-TEST",
		X:            0,
		Y:            0,
		Traits: []schema.Trait{
			{
				Symbol:      "MARKETPLACE",
				Name:        "Marketplace",
				Description: "A bustling marketplace",
			},
		},
	}
	gs.Waypoints[waypoint.Symbol] = waypoint

	// Add sample market
	market := &schema.Market{
		Symbol: "X1-TEST-A1",
		Exports: []schema.TradeGood{
			{
				Symbol:      "IRON",
				Name:        "Iron",
				Description: "Raw iron ore",
			},
		},
		Imports: []schema.TradeGood{
			{
				Symbol:      "FOOD",
				Name:        "Food",
				Description: "Nutritious food supplies",
			},
		},
	}
	gs.Markets[market.Symbol] = market

	// Initialize fuel prices
	gs.FuelPrices["X1-TEST-A1"] = 100

	// Initialize market prices
	gs.MarketPrices["X1-TEST-A1"] = map[string]int{
		"IRON": 50,
		"FOOD": 25,
	}
}

// Placeholder implementations for ship operations
func (m *MockServer) handleGetShip(w http.ResponseWriter, r *http.Request, shipSymbol string) {
	// Implementation would fetch and return ship data
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleShipOrbit(w http.ResponseWriter, r *http.Request, shipSymbol string) {
	// Implementation would change ship status to orbiting
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleShipDock(w http.ResponseWriter, r *http.Request, shipSymbol string) {
	// Implementation would change ship status to docked
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleShipNavigate(w http.ResponseWriter, r *http.Request, shipSymbol string) {
	// Implementation would handle navigation with fuel consumption and travel time
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleShipRefuel(w http.ResponseWriter, r *http.Request, shipSymbol string) {
	// Implementation would handle refueling with cost calculation
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handlePurchaseCargo(w http.ResponseWriter, r *http.Request, shipSymbol string) {
	// Implementation would handle cargo purchase with market price calculations
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleSellCargo(w http.ResponseWriter, r *http.Request, shipSymbol string) {
	// Implementation would handle cargo sales with market price calculations
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleSystemOperations(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle system and waypoint operations
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleGetContracts(w http.ResponseWriter, r *http.Request) {
	// Implementation would return available contracts
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}

func (m *MockServer) handleContractOperations(w http.ResponseWriter, r *http.Request) {
	// Implementation would handle contract accept/fulfill operations
	m.writeError(w, http.StatusNotImplemented, "Not implemented in basic version")
}