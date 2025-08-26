# SpaceTraders Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/JoeEdwardsCode/spacetraders-client.svg)](https://pkg.go.dev/github.com/JoeEdwardsCode/spacetraders-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/JoeEdwardsCode/spacetraders-client)](https://goreportcard.com/report/github.com/JoeEdwardsCode/spacetraders-client)

A comprehensive Go client library for the [SpaceTraders API](https://spacetraders.io/), a space trading game where you control automated ships, trade resources, and explore the galaxy.

## Features

- **Complete API Coverage**: Full implementation of SpaceTraders API v2
- **Zero External Dependencies**: Uses only Go standard library  
- **Thread-Safe**: Concurrent operations with proper synchronization
- **Rate Limiting**: Built-in compliance with API rate limits (2 req/sec, 30 burst)
- **Mock Server**: Comprehensive testing with realistic game logic simulation
- **Context Support**: Timeout and cancellation support for all operations
- **Comprehensive Testing**: Unit and integration tests with mock server

## Installation

```bash
go get github.com/JoeEdwardsCode/spacetraders-client
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "github.com/JoeEdwardsCode/spacetraders-client/pkg/client"
)

func main() {
    // Create a client with default configuration
    config := client.DefaultConfig()
    client, err := client.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()

    // Register a new agent
    resp, err := client.RegisterAgent(ctx, "MY_CALLSIGN", "COSMIC")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Registered agent: %s", resp.Agent.Symbol)
    log.Printf("Starting credits: %d", resp.Agent.Credits)
    log.Printf("Starting ship: %s", resp.Ship.Symbol)

    // Get your fleet
    ships, err := client.GetFleet(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Fleet size: %d ships", len(ships))
    for _, ship := range ships {
        log.Printf("- %s (%s) at %s", ship.Symbol, ship.Registration.Role, ship.Nav.WaypointSymbol)
    }
}
```

## Using with an Existing Token

```go
config := client.DefaultConfig()
config.Token = "your-existing-token"  // Set your existing token
client, err := client.New(config)
if err != nil {
    log.Fatal(err)
}

// You can also set the token later
client.SetToken("your-token")
```

## Project Structure

```
pkg/
├── client/          # Core API client
├── auth/            # Authentication handling
├── schema/          # Generated types from OpenAPI
├── endpoints/       # API endpoint implementations
├── transport/       # HTTP transport & rate limiting
├── cache/           # Client-side caching
├── fleet/           # Multi-ship management
└── mock/            # Mock server for testing
```

## Testing

```bash
# Run all tests
go test ./...

# Run with mock server
go test ./tests/integration/

# Run benchmarks
go test -bench=. ./tests/benchmarks/
```

## License

MIT License