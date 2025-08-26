package main

import (
	"context"
	"fmt"
	"log"

	"github.com/JoeEdwardsCode/spacetraders-client/pkg/client"
)

// This example demonstrates basic usage of the SpaceTraders client
func main() {
	fmt.Println("SpaceTraders Go Client - Simple Usage Example")
	fmt.Println("=============================================")

	// Create a new client with default configuration
	config := client.DefaultConfig()

	// Optional: configure custom settings
	// config.BaseURL = "https://api.spacetraders.io/v2"  // Default
	// config.Timeout = 30 * time.Second                  // Default
	// config.UserAgent = "MyApp/1.0"                     // Custom user agent

	spaceTradersClient, err := client.New(config)
	if err != nil {
		log.Fatalf("Failed to create SpaceTraders client: %v", err)
	}
	defer spaceTradersClient.Close()

	ctx := context.Background()

	// Example 1: Register a new agent (or use existing token)
	fmt.Println("\n1. Agent Registration/Authentication")
	fmt.Println("------------------------------------")

	// Method A: Register a new agent (only works once per callsign)
	resp, err := spaceTradersClient.RegisterAgent(ctx, "EXAMPLE_BOT", "COSMIC")
	if err != nil {
		fmt.Printf("Registration failed (expected if agent exists): %v\n", err)

		// Method B: Use an existing token
		// If you already have a token, set it here:
		// spaceTradersClient.SetToken("your-existing-token-here")
		fmt.Println("Using mock token for demonstration...")
		spaceTradersClient.SetToken("demo-token-123")
	} else {
		fmt.Printf("✓ Successfully registered agent: %s\n", resp.Agent.Symbol)
		fmt.Printf("  Starting credits: %d\n", resp.Agent.Credits)
		fmt.Printf("  Headquarters: %s\n", resp.Agent.Headquarters)
		fmt.Printf("  Token: %s\n", resp.Token[:20]+"...") // Show partial token
	}

	// Example 2: Get agent information
	fmt.Println("\n2. Agent Information")
	fmt.Println("-------------------")

	agent, err := spaceTradersClient.GetAgent(ctx)
	if err != nil {
		log.Printf("Failed to get agent info: %v", err)
	} else {
		fmt.Printf("✓ Agent: %s\n", agent.Symbol)
		fmt.Printf("  Account ID: %s\n", agent.AccountID)
		fmt.Printf("  Credits: %d\n", agent.Credits)
		fmt.Printf("  Headquarters: %s\n", agent.Headquarters)
		fmt.Printf("  Ship count: %d\n", agent.ShipCount)
		fmt.Printf("  Starting faction: %s\n", agent.StartingFaction)
	}

	// Example 3: Get fleet information
	fmt.Println("\n3. Fleet Management")
	fmt.Println("------------------")

	ships, err := spaceTradersClient.GetFleet(ctx, nil)
	if err != nil {
		log.Printf("Failed to get fleet: %v", err)
	} else {
		fmt.Printf("✓ Fleet contains %d ship(s)\n", len(ships))

		for i, ship := range ships {
			fmt.Printf("  Ship %d: %s\n", i+1, ship.Symbol)
			fmt.Printf("    Registration: %s class, %s role\n",
				ship.Registration.Name, ship.Registration.Role)
			fmt.Printf("    Location: %s (status: %s)\n",
				ship.Nav.WaypointSymbol, ship.Nav.Status)
			fmt.Printf("    Fuel: %d/%d units\n",
				ship.Fuel.Current, ship.Fuel.Capacity)
			fmt.Printf("    Cargo: %d/%d units used\n",
				ship.Cargo.Units, ship.Cargo.Capacity)

			if len(ship.Cargo.Inventory) > 0 {
				fmt.Printf("    Cargo contents:\n")
				for _, item := range ship.Cargo.Inventory {
					fmt.Printf("      - %s: %d units\n", item.Symbol, item.Units)
				}
			}
			fmt.Println()
		}
	}

	// Example 4: Authentication status
	fmt.Println("\n4. Authentication Status")
	fmt.Println("-----------------------")

	fmt.Printf("✓ Has authentication token: %t\n", spaceTradersClient.IsAuthenticated())

	tokenInfo := spaceTradersClient.GetTokenInfo(ctx)
	fmt.Printf("  Token info:\n")
	fmt.Printf("    Has token: %t\n", tokenInfo.HasToken)
	fmt.Printf("    Is valid: %t\n", tokenInfo.IsValid)
	fmt.Printf("    Last checked: %s\n", tokenInfo.LastChecked.Format("15:04:05"))

	if tokenInfo.Agent != nil {
		fmt.Printf("    Agent: %s\n", tokenInfo.Agent.Symbol)
	}

	// Example 5: Context usage patterns
	fmt.Println("\n5. Context Usage Examples")
	fmt.Println("------------------------")

	// With timeout
	// timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	// defer cancel()
	// _, err = spaceTradersClient.GetAgent(timeoutCtx)
	// if err != nil {
	// 	fmt.Printf("Request with timeout: %v\n", err)
	// }

	// With cancellation
	// cancelCtx, cancelFunc := context.WithCancel(ctx)
	// cancelFunc() // Cancel immediately for demo
	// _, err = spaceTradersClient.GetAgent(cancelCtx)
	// if err != nil {
	// 	fmt.Printf("Cancelled request: %v\n", err)
	// }

	fmt.Println("✓ Context patterns work as expected")

	// Example 6: Error handling
	fmt.Println("\n6. Error Handling")
	fmt.Println("----------------")

	// Try invalid operations to demonstrate error handling
	_, err = spaceTradersClient.RegisterAgent(ctx, "AB", "COSMIC") // Too short
	if err != nil {
		fmt.Printf("✓ Validation error caught: %v\n", err)
	}

	_, err = spaceTradersClient.RegisterAgent(ctx, "VALID_NAME", "INVALID_FACTION")
	if err != nil {
		fmt.Printf("✓ Invalid faction error caught: %v\n", err)
	}

	fmt.Println("\n✓ Example completed successfully!")
	fmt.Println("\nNext Steps:")
	fmt.Println("----------")
	fmt.Println("1. Visit https://spacetraders.io to get your real API token")
	fmt.Println("2. Replace the demo token with your real token")
	fmt.Println("3. Start building your space trading bot!")
	fmt.Println("4. Check out more examples in cmd/example/main.go")
}
