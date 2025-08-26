// Package schema contains all data types and structures for the SpaceTraders API
package schema

import (
	"time"
)

// Agent represents a SpaceTraders agent (player)
type Agent struct {
	AccountID       string `json:"accountId"`
	Symbol          string `json:"symbol"`
	Headquarters    string `json:"headquarters"`
	Credits         int64  `json:"credits"`
	StartingFaction string `json:"startingFaction"`
	ShipCount       int    `json:"shipCount"`
}

// Ship represents a SpaceTraders ship
type Ship struct {
	Symbol       string       `json:"symbol"`
	Registration Registration `json:"registration"`
	Nav          Navigation   `json:"nav"`
	Crew         Crew         `json:"crew"`
	Frame        Frame        `json:"frame"`
	Reactor      Reactor      `json:"reactor"`
	Engine       Engine       `json:"engine"`
	Modules      []Module     `json:"modules"`
	Mounts       []Mount      `json:"mounts"`
	Cargo        Cargo        `json:"cargo"`
	Fuel         Fuel         `json:"fuel"`
}

// Registration holds ship registration information
type Registration struct {
	Name          string `json:"name"`
	FactionSymbol string `json:"factionSymbol"`
	Role          string `json:"role"`
}

// Navigation contains ship navigation information
type Navigation struct {
	SystemSymbol   string         `json:"systemSymbol"`
	WaypointSymbol string         `json:"waypointSymbol"`
	Route          Route          `json:"route"`
	Status         string         `json:"status"`
	FlightMode     string         `json:"flightMode"`
}

// Route represents a navigation route
type Route struct {
	Destination   RouteWaypoint `json:"destination"`
	Origin        RouteWaypoint `json:"origin"`
	DepartureTime time.Time     `json:"departureTime"`
	Arrival       time.Time     `json:"arrival"`
}

// RouteWaypoint represents a waypoint in a route
type RouteWaypoint struct {
	Symbol       string `json:"symbol"`
	Type         string `json:"type"`
	SystemSymbol string `json:"systemSymbol"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
}

// Crew represents ship crew information
type Crew struct {
	Current  int    `json:"current"`
	Required int    `json:"required"`
	Capacity int    `json:"capacity"`
	Rotation string `json:"rotation"`
	Morale   int    `json:"morale"`
	Wages    int    `json:"wages"`
}

// Frame represents ship frame information
type Frame struct {
	Symbol         string              `json:"symbol"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Condition      int                 `json:"condition"`
	Integrity      int                 `json:"integrity"`
	ModuleSlots    int                 `json:"moduleSlots"`
	MountingPoints int                 `json:"mountingPoints"`
	FuelCapacity   int                 `json:"fuelCapacity"`
	Requirements   ShipRequirements    `json:"requirements"`
}

// Reactor represents ship reactor information
type Reactor struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    int              `json:"condition"`
	Integrity    int              `json:"integrity"`
	PowerOutput  int              `json:"powerOutput"`
	Requirements ShipRequirements `json:"requirements"`
}

// Engine represents ship engine information
type Engine struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    int              `json:"condition"`
	Integrity    int              `json:"integrity"`
	Speed        int              `json:"speed"`
	Requirements ShipRequirements `json:"requirements"`
}

// Module represents a ship module
type Module struct {
	Symbol       string           `json:"symbol"`
	Capacity     *int             `json:"capacity,omitempty"`
	Range        *int             `json:"range,omitempty"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Requirements ShipRequirements `json:"requirements"`
}

// Mount represents a ship mount
type Mount struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Strength     *int             `json:"strength,omitempty"`
	Requirements ShipRequirements `json:"requirements"`
}

// ShipRequirements represents requirements for ship components
type ShipRequirements struct {
	Power *int `json:"power,omitempty"`
	Crew  *int `json:"crew,omitempty"`
	Slots *int `json:"slots,omitempty"`
}

// Cargo represents ship cargo information
type Cargo struct {
	Capacity  int           `json:"capacity"`
	Units     int           `json:"units"`
	Inventory []CargoItem   `json:"inventory"`
}

// CargoItem represents an item in cargo
type CargoItem struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Units       int    `json:"units"`
}

// Fuel represents ship fuel information
type Fuel struct {
	Current  int       `json:"current"`
	Capacity int       `json:"capacity"`
	Consumed *FuelUsed `json:"consumed,omitempty"`
}

// FuelUsed represents fuel consumption data
type FuelUsed struct {
	Amount    int       `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

// Contract represents a SpaceTraders contract
type Contract struct {
	ID               string           `json:"id"`
	FactionSymbol    string           `json:"factionSymbol"`
	Type             string           `json:"type"`
	Terms            ContractTerms    `json:"terms"`
	Accepted         bool             `json:"accepted"`
	Fulfilled        bool             `json:"fulfilled"`
	Expiration       time.Time        `json:"expiration"`
	DeadlineToAccept *time.Time       `json:"deadlineToAccept,omitempty"`
}

// ContractTerms represents contract terms
type ContractTerms struct {
	Deadline time.Time             `json:"deadline"`
	Payment  ContractPayment       `json:"payment"`
	Deliver  []ContractDeliverGood `json:"deliver,omitempty"`
}

// ContractPayment represents contract payment information
type ContractPayment struct {
	OnAccepted  int `json:"onAccepted"`
	OnFulfilled int `json:"onFulfilled"`
}

// ContractDeliverGood represents a good to be delivered for a contract
type ContractDeliverGood struct {
	TradeSymbol       string `json:"tradeSymbol"`
	DestinationSymbol string `json:"destinationSymbol"`
	UnitsRequired     int    `json:"unitsRequired"`
	UnitsFulfilled    int    `json:"unitsFulfilled"`
}

// Market represents a SpaceTraders market
type Market struct {
	Symbol       string      `json:"symbol"`
	Exports      []TradeGood `json:"exports"`
	Imports      []TradeGood `json:"imports"`
	Exchange     []TradeGood `json:"exchange"`
	Transactions []Transaction `json:"transactions,omitempty"`
	TradeGoods   []TradeGood `json:"tradeGoods,omitempty"`
}

// TradeGood represents a tradeable good
type TradeGood struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Type          string `json:"type,omitempty"`
	TradeVolume   *int   `json:"tradeVolume,omitempty"`
	Supply        *string `json:"supply,omitempty"`
	PurchasePrice *int   `json:"purchasePrice,omitempty"`
	SellPrice     *int   `json:"sellPrice,omitempty"`
}

// Transaction represents a market transaction
type Transaction struct {
	WaypointSymbol string    `json:"waypointSymbol"`
	ShipSymbol     string    `json:"shipSymbol"`
	TradeSymbol    string    `json:"tradeSymbol"`
	Type           string    `json:"type"`
	Units          int       `json:"units"`
	PricePerUnit   int       `json:"pricePerUnit"`
	TotalPrice     int       `json:"totalPrice"`
	Timestamp      time.Time `json:"timestamp"`
}

// System represents a SpaceTraders system
type System struct {
	Symbol       string     `json:"symbol"`
	SectorSymbol string     `json:"sectorSymbol"`
	Type         string     `json:"type"`
	X            int        `json:"x"`
	Y            int        `json:"y"`
	Waypoints    []Waypoint `json:"waypoints"`
	Factions     []Faction  `json:"factions"`
}

// Waypoint represents a waypoint in a system
type Waypoint struct {
	Symbol       string    `json:"symbol"`
	Type         string    `json:"type"`
	SystemSymbol string    `json:"systemSymbol"`
	X            int       `json:"x"`
	Y            int       `json:"y"`
	Orbitals     []Orbital `json:"orbitals"`
	Traits       []Trait   `json:"traits"`
	Modifiers    []Modifier `json:"modifiers,omitempty"`
	Chart        *Chart    `json:"chart,omitempty"`
	Faction      *Faction  `json:"faction,omitempty"`
}

// Orbital represents an orbital body
type Orbital struct {
	Symbol string `json:"symbol"`
}

// Trait represents a waypoint trait
type Trait struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Modifier represents a waypoint modifier
type Modifier struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Chart represents waypoint chart information
type Chart struct {
	WaypointSymbol *string   `json:"waypointSymbol,omitempty"`
	SubmittedBy    *string   `json:"submittedBy,omitempty"`
	SubmittedOn    *time.Time `json:"submittedOn,omitempty"`
}

// Faction represents a SpaceTraders faction
type Faction struct {
	Symbol       string     `json:"symbol"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Headquarters string     `json:"headquarters"`
	Traits       []FactionTrait `json:"traits"`
	IsRecruiting bool       `json:"isRecruiting"`
}

// FactionTrait represents a faction trait
type FactionTrait struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Survey represents a mining survey
type Survey struct {
	Signature   string        `json:"signature"`
	Symbol      string        `json:"symbol"`
	Deposits    []SurveyDeposit `json:"deposits"`
	Expiration  time.Time     `json:"expiration"`
	Size        string        `json:"size"`
}

// SurveyDeposit represents a deposit found in a survey
type SurveyDeposit struct {
	Symbol string `json:"symbol"`
}

// Extraction represents a resource extraction result
type Extraction struct {
	ShipSymbol string      `json:"shipSymbol"`
	Yield      ExtractionYield `json:"yield"`
}

// ExtractionYield represents the yield from extraction
type ExtractionYield struct {
	Symbol string `json:"symbol"`
	Units  int    `json:"units"`
}

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Data interface{} `json:"data"`
	Meta *Meta       `json:"meta,omitempty"`
}

// Meta represents pagination and response metadata
type Meta struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// APIError represents an API error response
type APIError struct {
	Message string            `json:"message"`
	Code    int               `json:"code"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// Registration request/response types

// RegisterAgentRequest represents a request to register a new agent
type RegisterAgentRequest struct {
	Symbol  string `json:"symbol"`
	Faction string `json:"faction"`
	Email   string `json:"email,omitempty"`
}

// RegisterAgentResponse represents the response from agent registration
type RegisterAgentResponse struct {
	Agent    Agent    `json:"agent"`
	Contract Contract `json:"contract"`
	Faction  Faction  `json:"faction"`
	Ship     Ship     `json:"ship"`
	Token    string   `json:"token"`
}

// Common request types

// PaginationOptions represents pagination query parameters
type PaginationOptions struct {
	Page  *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}

// NavigateShipRequest represents a request to navigate a ship
type NavigateShipRequest struct {
	WaypointSymbol string `json:"waypointSymbol"`
}

// PurchaseCargoRequest represents a request to purchase cargo
type PurchaseCargoRequest struct {
	Symbol string `json:"symbol"`
	Units  int    `json:"units"`
}

// SellCargoRequest represents a request to sell cargo
type SellCargoRequest struct {
	Symbol string `json:"symbol"`
	Units  int    `json:"units"`
}