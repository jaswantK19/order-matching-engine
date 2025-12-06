# Order Matching Engine

Hey, this is my implementation of the order matching engine. It's built in Go and I tried to keep it pretty simple but fast.

## How to Run It

First, make sure you have Go installed.

### Run the Server
```bash
go run cmd/server/main.go
```
It runs on port 8080.

### Run Tests
I wrote a bunch of tests to make sure things work right.
```bash
go test ./...
```

### Run Benchmarks
If u wanna see how fast it is:
```bash
go test -bench=. ./internal/engine
```

### Run the Compliance Load Test
This is the big one that checks the 60s load requirement.
```bash
go test -v -timeout 70s -run TestComplianceRequirements ./internal/engine
```

## Benchmarks & Compliance

I ran the full **60-second compliance load test** with 100 concurrent clients on my machine (Apple M5).

Here are the results:

```text
=== Compliance Check Results ===
Duration:       1m0.001s
Total Orders:   27,532,316
Throughput:     458,863.55 orders/sec (Target: >30,000)
Latency P50:    89.46µs (Target: <10ms)
Latency P99:    1.16ms (Target: <50ms)
Latency P99.9:  13.03ms (Target: <100ms)
Memory Usage:   ~9 GB (Allocated for ~27M active orders)
================================
```

It crushed the requirements!
- **Throughput**: ~15x the target.
- **Latency**: P99 is way under the 50ms limit.

### Simulated Standard Hardware (Fairness Check)
To simulate a standard 4-core machine and prove concurrency safety, I ran the test with `GOMAXPROCS=4` and the `-race` detector enabled (which significantly slows down execution).

```bash
GOMAXPROCS=4 go test -race -v -run TestComplianceRequirements ./internal/engine
```

**Results:**
- **Throughput**: ~150,598 orders/sec (Still 5x the target!)
- **Latency P99**: ~1.63ms
- **Race Detection**: Passed (No race conditions found).

## Profiling & Tools
I've integrated standard Go tooling support:
- **pprof**: Enabled at `/debug/pprof/`. You can profile the running server using `go tool pprof http://localhost:8080/debug/pprof/profile`.
- **Race Detector**: The system is race-free. Run tests with `go test -race ./...`.
- **Benchstat**: Benchmarks are standard Go benchmarks, compatible with `benchstat`.


👉 **[Read the Explanation & Strategy](EXPLANATION.md)**
