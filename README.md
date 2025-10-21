# Cloud Computing Learning Projects

A collection of Go projects demonstrating cloud computing patterns and distributed systems concepts.

## Projects

### 🔄 [Token Caching](./token-caching/)
- In-memory token cache with TTL
- Concurrent readers/writers using RWMutex
- Background cleanup goroutine
- **Concepts:** Caching, Concurrency, Goroutines

### ⚡ [Circuit Breaker](./circuit-breaker/)
- HTTP circuit breaker pattern implementation
- Fail-fast mechanism with timeout recovery
- Fallback responses for resilience
- **Concepts:** Fault Tolerance, Resilience Patterns

### 🔗 [Context Demo](./context-demo/)
- Context usage patterns
- Timeout and cancellation handling
- **Concepts:** Request Context, Timeouts

## How to Run

Each project has its own README with specific instructions. Generally:

```bash
cd project-folder
go run main.go
```

## Learning Goals
- Distributed systems patterns
- Concurrent programming in Go
- Microservices resilience
- Cloud-native development