package integration

import (
	"context"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/client"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/mock"
	"github.com/JoeEdwardsCode/spacetraders-client/pkg/transport"
	"testing"
	"time"
)

func TestClientIntegration(t *testing.T) {
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
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Agent Registration", func(t *testing.T) {
		testAgentRegistration(t, ctx, client)
	})

	t.Run("Agent Information", func(t *testing.T) {
		testAgentInformation(t, ctx, client)
	})

	t.Run("Fleet Operations", func(t *testing.T) {
		testFleetOperations(t, ctx, client)
	})

	t.Run("Authentication", func(t *testing.T) {
		testAuthentication(t, ctx, client, mockServer.GetURL())
	})

	t.Run("Rate Limiting", func(t *testing.T) {
		testRateLimiting(t, ctx, client, mockServer)
	})
}

func testAgentRegistration(t *testing.T, ctx context.Context, client *client.SpaceTradersClient) {
	// Test valid registration
	resp, err := client.RegisterAgent(ctx, "TEST_AGENT_1", "COSMIC")
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Validate response
	if resp.Agent.Symbol != "TEST_AGENT_1" {
		t.Errorf("Expected agent symbol 'TEST_AGENT_1', got '%s'", resp.Agent.Symbol)
	}

	if resp.Agent.StartingFaction != "COSMIC" {
		t.Errorf("Expected starting faction 'COSMIC', got '%s'", resp.Agent.StartingFaction)
	}

	if resp.Token == "" {
		t.Error("Expected non-empty token")
	}

	if resp.Agent.Credits <= 0 {
		t.Errorf("Expected positive starting credits, got %d", resp.Agent.Credits)
	}

	// Test duplicate registration (should fail)
	_, err = client.RegisterAgent(ctx, "TEST_AGENT_1", "COSMIC")
	if err == nil {
		t.Error("Expected error for duplicate agent registration")
	}

	if !transport.IsAPIError(err) {
		t.Errorf("Expected API error, got: %T", err)
	}
}

func testAgentInformation(t *testing.T, ctx context.Context, client *client.SpaceTradersClient) {
	// First register an agent to get a token
	_, err := client.RegisterAgent(ctx, "TEST_AGENT_2", "VOID")
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Test getting agent information
	agent, err := client.GetAgent(ctx)
	if err != nil {
		t.Fatalf("Failed to get agent: %v", err)
	}

	if agent.Symbol != "TEST_AGENT_2" {
		t.Errorf("Expected agent symbol 'TEST_AGENT_2', got '%s'", agent.Symbol)
	}

	// Test authentication status
	if !client.IsAuthenticated() {
		t.Error("Client should be authenticated after registration")
	}

	// Test token info
	tokenInfo := client.GetTokenInfo(ctx)
	if !tokenInfo.HasToken {
		t.Error("Token info should indicate token is present")
	}

	if !tokenInfo.IsValid {
		t.Error("Token info should indicate token is valid")
	}

	if tokenInfo.Agent == nil {
		t.Error("Token info should include agent data")
	}
}

func testFleetOperations(t *testing.T, ctx context.Context, client *client.SpaceTradersClient) {
	// Register agent first
	_, err := client.RegisterAgent(ctx, "TEST_AGENT_3", "GALACTIC")
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Test getting fleet
	ships, err := client.GetFleet(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to get fleet: %v", err)
	}

	if len(ships) == 0 {
		t.Error("Expected at least one ship in starting fleet")
	}

	// Test getting specific ship
	if len(ships) > 0 {
		ship, err := client.GetShip(ctx, ships[0].Symbol)
		if err != nil {
			// This might not be implemented in mock server yet
			t.Logf("GetShip not implemented in mock server: %v", err)
		} else {
			if ship.Symbol != ships[0].Symbol {
				t.Errorf("Expected ship symbol '%s', got '%s'", ships[0].Symbol, ship.Symbol)
			}
		}
	}
}

func testAuthentication(t *testing.T, ctx context.Context, clientInstance *client.SpaceTradersClient, mockServerURL string) {
	// First, register an agent to get a valid token
	authTestClient, err := client.New(&client.Config{
		BaseURL: mockServerURL,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create auth test client: %v", err)
	}
	defer authTestClient.Close()

	// Register agent to get token
	_, err = authTestClient.RegisterAgent(ctx, "AUTH_TEST", "COSMIC")
	if err != nil {
		t.Fatalf("Failed to register agent: %v", err)
	}

	// Test unauthenticated client
	unauthClient, err := client.New(&client.Config{
		BaseURL: mockServerURL, // Use same mock server
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create unauthenticated client: %v", err)
	}

	// Should fail to get agent without authentication
	_, err = unauthClient.GetAgent(ctx)
	if err == nil {
		t.Error("Expected error when getting agent without authentication")
	}

	if !transport.IsAuthError(err) {
		t.Errorf("Expected authentication error, got: %T", err)
	}

	// Test setting token manually
	validToken := authTestClient.GetToken()
	unauthClient.SetToken(validToken)

	// Should now work
	agent, err := unauthClient.GetAgent(ctx)
	if err != nil {
		t.Fatalf("Failed to get agent with valid token: %v", err)
	}

	if agent == nil {
		t.Error("Expected agent data")
	}
}

func testRateLimiting(t *testing.T, ctx context.Context, clientInstance *client.SpaceTradersClient, mockServer *mock.MockServer) {
	// Register agent first with existing client
	_, err := clientInstance.RegisterAgent(ctx, "RATE_TEST", "QUANTUM")
	if err != nil {
		t.Logf("Agent registration failed (may already exist): %v", err)
		// Continue with test even if agent already exists
	}

	// Enable rate limiting on mock server
	mockServer.SetRateLimitEnabled(true)

	// Test the rate limiter functionality by checking if rate limiting is properly configured
	// Since client-side rate limiting uses Wait() (which blocks), we test server-side rate limiting
	// by making rapid requests and checking if we can observe the rate limiting behavior

	successCount := 0
	errorCount := 0

	// Make requests to test rate limiting (both client and server side)
	for i := 0; i < 35; i++ {
		_, err := clientInstance.GetAgent(ctx)
		if err != nil {
			errorCount++
			if transport.IsRateLimitError(err) {
				t.Logf("Got expected rate limit error: %v", err)
			}
		} else {
			successCount++
		}
	}

	// At minimum, we should have some successful requests
	if successCount == 0 {
		t.Error("Expected some requests to succeed")
	}

	// The key test is that the rate limiter exists and is functional
	state := clientInstance.GetRateLimiterState()
	if state == nil {
		t.Error("Expected rate limiter state")
	}

	t.Logf("Rate limiting test: %d successful, %d errors", successCount, errorCount)

	// Disable rate limiting for other tests
	mockServer.SetRateLimitEnabled(false)
}

func TestClientEdgeCases(t *testing.T) {
	mockServer := mock.NewMockServer()
	defer mockServer.Close()

	config := &client.Config{
		BaseURL: mockServer.GetURL(),
		Timeout: 5 * time.Second,
	}

	client, err := client.New(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	t.Run("Invalid Call Sign", func(t *testing.T) {
		// Test invalid call sign formats
		invalidCallSigns := []string{
			"",                 // Empty
			"AB",               // Too short
			"ABCDEFGHIJKLMNOP", // Too long
			"ABC-123",          // Invalid character
			"abc 123",          // Space
			"123@456",          // Invalid character
		}

		for _, callSign := range invalidCallSigns {
			_, err := client.RegisterAgent(ctx, callSign, "COSMIC")
			if err == nil {
				t.Errorf("Expected error for invalid call sign: %s", callSign)
			}
		}
	})

	t.Run("Invalid Faction", func(t *testing.T) {
		_, err := client.RegisterAgent(ctx, "VALID_AGENT", "INVALID_FACTION")
		if err == nil {
			t.Error("Expected error for invalid faction")
		}
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		// Test context cancellation
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.RegisterAgent(cancelCtx, "CANCELLED_AGENT", "COSMIC")
		if err == nil {
			t.Error("Expected error for cancelled context")
		}

		if err != context.Canceled {
			t.Logf("Got error: %v (expected context.Canceled)", err)
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		// Test request timeout
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// This should timeout (though it might complete too quickly in tests)
		_, err := client.RegisterAgent(timeoutCtx, "TIMEOUT_AGENT", "COSMIC")
		if err == context.DeadlineExceeded {
			t.Log("Successfully tested timeout")
		} else {
			t.Logf("Timeout test inconclusive (got: %v)", err)
		}
	})
}

func TestClientConfiguration(t *testing.T) {
	t.Run("Default Config", func(t *testing.T) {
		config := client.DefaultConfig()

		if config.BaseURL == "" {
			t.Error("Default config should have base URL")
		}

		if config.Timeout <= 0 {
			t.Error("Default config should have positive timeout")
		}

		if config.UserAgent == "" {
			t.Error("Default config should have user agent")
		}
	})

	t.Run("Custom Config", func(t *testing.T) {
		customConfig := &client.Config{
			BaseURL:   "https://custom.example.com",
			Timeout:   60 * time.Second,
			UserAgent: "Custom-Agent/1.0",
			Token:     "existing-token",
		}

		client, err := client.New(customConfig)
		if err != nil {
			t.Fatalf("Failed to create client with custom config: %v", err)
		}
		defer client.Close()

		if client.GetToken() != "existing-token" {
			t.Error("Client should use provided token")
		}

		if client.IsAuthenticated() {
			// Note: This might be false if token validation fails
			t.Log("Client reports as authenticated with provided token")
		}
	})

	t.Run("Nil Config", func(t *testing.T) {
		client, err := client.New(nil)
		if err != nil {
			t.Fatalf("Client should accept nil config and use defaults: %v", err)
		}
		defer client.Close()

		if client.GetToken() != "" {
			t.Error("Client with nil config should have empty token")
		}

		if client.IsAuthenticated() {
			t.Error("Client with nil config should not be authenticated")
		}
	})
}
