package main

import (
	"context"
	"fmt"
	"log"
	"spacetraders-client/pkg/client"
	"time"
)

func main() {
	fmt.Println("SpaceTraders Golang Client Example")
	fmt.Println("===================================")

	// Create client with default configuration
	config := client.DefaultConfig()
	client, err := client.New(config)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 1: Register a new agent
	fmt.Println("\n1. Registering new agent...")
	resp, err := client.RegisterAgent(ctx, "EXAMPLE_AGENT", "COSMIC")
	if err != nil {
		log.Printf("Registration failed (expected on real API): %v", err)
		
		// For demo purposes, set a mock token
		client.SetToken("demo-token-for-testing")
	} else {
		fmt.Printf("✓ Agent registered: %s\n", resp.Agent.Symbol)
		fmt.Printf("  Starting credits: %d\n", resp.Agent.Credits)
		fmt.Printf("  Headquarters: %s\n", resp.Agent.Headquarters)
		fmt.Printf("  Starting ship: %s\n", resp.Ship.Symbol)
		fmt.Printf("  First contract: %s\n", resp.Contract.ID)
	}

	// Example 2: Get agent information
	fmt.Println("\n2. Getting agent information...")
	agent, err := client.GetAgent(ctx)
	if err != nil {
		log.Printf("Failed to get agent: %v", err)
	} else {
		fmt.Printf("✓ Agent: %s\n", agent.Symbol)
		fmt.Printf("  Credits: %d\n", agent.Credits)
		fmt.Printf("  Ship count: %d\n", agent.ShipCount)
	}

	// Example 3: Get fleet
	fmt.Println("\n3. Getting fleet information...")
	ships, err := client.GetFleet(ctx, nil)
	if err != nil {
		log.Printf("Failed to get fleet: %v", err)
	} else {
		fmt.Printf("✓ Fleet size: %d ships\n", len(ships))
		for i, ship := range ships {
			fmt.Printf("  Ship %d: %s (Role: %s, Status: %s)\n", 
				i+1, ship.Symbol, ship.Registration.Role, ship.Nav.Status)
			fmt.Printf("    Location: %s\n", ship.Nav.WaypointSymbol)
			fmt.Printf("    Fuel: %d/%d\n", ship.Fuel.Current, ship.Fuel.Capacity)
			fmt.Printf("    Cargo: %d/%d\n", ship.Cargo.Units, ship.Cargo.Capacity)
		}
	}

	// Example 4: Authentication status
	fmt.Println("\n4. Checking authentication...")
	fmt.Printf("✓ Authenticated: %t\n", client.IsAuthenticated())
	
	tokenInfo := client.GetTokenInfo(ctx)
	fmt.Printf("  Has token: %t\n", tokenInfo.HasToken)
	fmt.Printf("  Token valid: %t\n", tokenInfo.IsValid)
	fmt.Printf("  Last checked: %s\n", tokenInfo.LastChecked.Format(time.RFC3339))

	// Example 5: Rate limiter status
	fmt.Println("\n5. Rate limiter status...")
	state := client.GetRateLimiterState()
	fmt.Printf("✓ Rate limiter state: %+v\n", state)

	// Example 6: Demonstrate error handling
	fmt.Println("\n6. Error handling examples...")
	
	// Try to register with invalid call sign
	_, err = client.RegisterAgent(ctx, "A", "COSMIC") // Too short
	if err != nil {
		fmt.Printf("✓ Expected validation error: %v\n", err)
	}

	// Try to register with invalid faction
	_, err = client.RegisterAgent(ctx, "VALID_NAME", "INVALID_FACTION")
	if err != nil {
		fmt.Printf("✓ Expected faction error: %v\n", err)
	}

	fmt.Println("\n7. Advanced features (when implemented)...")
	
	// These would be available in the full implementation:
	
	// Market operations
	fmt.Println("  - Market data retrieval")
	fmt.Println("  - Cargo trading")
	
	// Contract operations  
	fmt.Println("  - Contract management")
	fmt.Println("  - Contract fulfillment")
	
	// Navigation
	fmt.Println("  - Ship navigation")
	fmt.Println("  - System exploration")
	
	// Mining
	fmt.Println("  - Resource surveys")
	fmt.Println("  - Resource extraction")

	fmt.Println("\n8. Usage patterns...")
	
	// Show timeout context usage
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	_, err = client.GetAgent(timeoutCtx)
	if err != nil {
		fmt.Printf("  Timeout example: %v\n", err)
	}

	// Show cancellation context usage
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	cancelFunc() // Cancel immediately
	
	_, err = client.GetAgent(cancelCtx)
	if err != nil {
		fmt.Printf("  Cancellation example: %v\n", err)
	}

	fmt.Println("\n✓ Example completed successfully!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Get your real SpaceTraders API token from https://spacetraders.io")
	fmt.Println("  2. Set your token: client.SetToken(\"your-real-token\")")
	fmt.Println("  3. Start building your space trading empire!")
}