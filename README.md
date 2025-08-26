# SpaceTraders Golang API Client

A comprehensive Golang API client for the SpaceTraders API with complete functionality, built-in rate limiting, and comprehensive testing.

## Features

- **Complete API Coverage**: Generated from OpenAPI specification
- **Zero External Dependencies**: Uses only Go standard library
- **Smart Rate Limiting**: Built-in compliance with API limits (2 req/sec, 30 burst)
- **Fleet Management**: Multi-ship coordination and optimization
- **Mock Server**: Comprehensive testing with business logic simulation
- **Caching Layer**: Intelligent caching with TTL management
- **Resilience**: Circuit breakers, retry logic, and graceful degradation

## Quick Start

```go
package main

import (
    "log"
    "spacetraders-client/pkg/client"
)

func main() {
    config := client.DefaultConfig()
    client, err := client.New(config)
    if err != nil {
        log.Fatal(err)
    }

    // Register a new agent
    agent, err := client.RegisterAgent("MY_CALLSIGN", "COSMIC")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Registered agent: %s", agent.Symbol)
}
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