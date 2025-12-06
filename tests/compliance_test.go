package tests

import (
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jaswantK19/order-matching-engine/internal/engine"
	"github.com/jaswantK19/order-matching-engine/internal/models"
)

func TestComplianceRequirements(t *testing.T) {
	const (
		duration       = 60 * time.Second
		numClients     = 100
		minThroughput  = 30000.0
		maxP50Latency  = 10 * time.Millisecond
		maxP99Latency  = 50 * time.Millisecond
		maxP999Latency = 100 * time.Millisecond
	)

	eng := engine.NewEngine()
	eng.Start()

	var (
		wg          sync.WaitGroup
		totalOrders int64
		startSignal = make(chan struct{})
		latencies   = make([][]time.Duration, numClients)
		symbols     = []string{"AAPL", "GOOG", "TSLA", "AMZN", "MSFT"}
	)

	for i := 0; i < numClients; i++ {
		latencies[i] = make([]time.Duration, 0, 50000)
	}

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			<-startSignal

			endTime := time.Now().Add(duration)
			localLatencies := latencies[clientID]
			localCount := 0

			for time.Now().Before(endTime) {
				sym := symbols[rand.Intn(len(symbols))]
				side := models.SideBuy
				if rand.Intn(2) == 0 {
					side = models.SideSell
				}
				price := int64(10000 + rand.Intn(100))
				qty := int64(1 + rand.Intn(100))

				o := &models.Order{
					ID:        fmt.Sprintf("load-%d-%d", clientID, localCount),
					Symbol:    sym,
					Side:      side,
					Type:      models.TypeLimit,
					Price:     price,
					Quantity:  qty,
					Timestamp: time.Now().UnixMilli(),
				}

				start := time.Now()
				eng.SubmitOrder(o)
				latency := time.Since(start)

				localLatencies = append(localLatencies, latency)
				localCount++
			}
			latencies[clientID] = localLatencies
			atomic.AddInt64(&totalOrders, int64(localCount))
		}(i)
	}

	fmt.Printf("Starting %d concurrent clients for %v load test...\n", numClients, duration)
	time.Sleep(1 * time.Second)

	testStart := time.Now()
	close(startSignal)
	wg.Wait()
	testDuration := time.Since(testStart)

	// Analysis
	var allLatencies []time.Duration
	for _, l := range latencies {
		allLatencies = append(allLatencies, l...)
	}
	sort.Slice(allLatencies, func(i, j int) bool {
		return allLatencies[i] < allLatencies[j]
	})

	count := len(allLatencies)
	if count == 0 {
		t.Fatal("No orders processed")
	}

	p50 := allLatencies[count*50/100]
	p99 := allLatencies[count*99/100]
	p999 := allLatencies[count*999/1000]

	throughput := float64(count) / testDuration.Seconds()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	fmt.Printf("\n=== Compliance Check Results ===\n")
	fmt.Printf("Duration:       %v\n", testDuration)
	fmt.Printf("Total Orders:   %d\n", count)
	fmt.Printf("Throughput:     %.2f orders/sec (Target: >%.0f)\n", throughput, minThroughput)
	fmt.Printf("Latency P50:    %v (Target: <%v)\n", p50, maxP50Latency)
	fmt.Printf("Latency P99:    %v (Target: <%v)\n", p99, maxP99Latency)
	fmt.Printf("Latency P99.9:  %v (Target: <%v)\n", p999, maxP999Latency)
	fmt.Printf("Memory Usage:   %.2f MB (Alloc) / %.2f MB (Sys)\n", allocMB, sysMB)
	fmt.Printf("================================\n")

	// Assertions
	if throughput < minThroughput {
		t.Errorf("FAIL: Throughput too low. Got %.2f, want >%.0f", throughput, minThroughput)
	}
	if p50 > maxP50Latency {
		t.Errorf("FAIL: P50 Latency too high. Got %v, want <%v", p50, maxP50Latency)
	}
	if p99 > maxP99Latency {
		t.Errorf("FAIL: P99 Latency too high. Got %v, want <%v", p99, maxP99Latency)
	}
	if p999 > maxP999Latency {
		t.Errorf("FAIL: P99.9 Latency too high. Got %v, want <%v", p999, maxP999Latency)
	}
}
