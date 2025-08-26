package main

import (
	"context"
	"fmt"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/client"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/mock"
	"log"
	"time"
)

func main() {
	fmt.Println("SpaceTraders Mock Server Test")
	fmt.Println("=============================")

	// Start mock server
	mockServer := mock.NewMockServer()
	defer mockServer.Close()

	// Create client with mock server URL
	config := &client.Config{
		BaseURL: mockServer.GetURL(),
		Timeout: 10 * time.Second,
	}

	client, err := client.New(config)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}
	defer client.Close()

	ctx := context.Background()

	fmt.Printf("Mock server running at: %s\n", mockServer.GetURL())

	// Test 1: Register an agent
	fmt.Println("\n1. Testing agent registration...")
	resp, err := client.RegisterAgent(ctx, "TEST_AGENT", "COSMIC")
	if err != nil {
		log.Printf("Registration failed: %v", err)
	} else {
		fmt.Printf("✓ Agent registered successfully!\n")
		fmt.Printf("  Symbol: %s\n", resp.Agent.Symbol)
		fmt.Printf("  Credits: %d\n", resp.Agent.Credits)
		fmt.Printf("  Token: %s\n", resp.Token[:20]+"...")
		fmt.Printf("  Ship: %s\n", resp.Ship.Symbol)
	}

	// Test 2: Get agent info
	fmt.Println("\n2. Testing get agent...")
	if client.IsAuthenticated() {
		agent, err := client.GetAgent(ctx)
		if err != nil {
			log.Printf("Get agent failed: %v", err)
		} else {
			fmt.Printf("✓ Agent retrieved successfully!\n")
			fmt.Printf("  Symbol: %s\n", agent.Symbol)
			fmt.Printf("  Credits: %d\n", agent.Credits)
		}
	} else {
		fmt.Printf("❌ Client not authenticated\n")
	}

	// Test 3: Get fleet
	fmt.Println("\n3. Testing get fleet...")
	ships, err := client.GetFleet(ctx, nil)
	if err != nil {
		log.Printf("Get fleet failed: %v", err)
	} else {
		fmt.Printf("✓ Fleet retrieved successfully!\n")
		fmt.Printf("  Ships: %d\n", len(ships))
		for i, ship := range ships {
			fmt.Printf("  Ship %d: %s\n", i+1, ship.Symbol)
		}
	}

	// Test 4: Rate limiting (disable first to see it work)
	fmt.Println("\n4. Testing without rate limiting...")
	mockServer.SetRateLimitEnabled(false)

	successCount := 0
	for i := 0; i < 5; i++ {
		_, err := client.GetAgent(ctx)
		if err == nil {
			successCount++
		}
	}
	fmt.Printf("✓ Made 5 requests successfully: %d succeeded\n", successCount)

	// Test 5: Enable rate limiting
	fmt.Println("\n5. Testing with rate limiting...")
	mockServer.SetRateLimitEnabled(true)

	successCount = 0
	limitedCount := 0
	for i := 0; i < 35; i++ {
		_, err := client.GetAgent(ctx)
		if err != nil {
			limitedCount++
		} else {
			successCount++
		}
	}
	fmt.Printf("✓ Rate limiting test: %d succeeded, %d limited\n", successCount, limitedCount)

	fmt.Println("\n✅ Mock server test completed!")
}
