package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/jaswantK19/order-matching-engine/internal/models"
)

// BenchmarkConcurrent simulates 100 concurrent clients
// Run with: go test -bench=. ./internal/engine -benchtime=5s
func BenchmarkConcurrent(b *testing.B) {
	eng := NewEngine()
	eng.Start()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			side := models.SideBuy
			if i%2 == 0 {
				side = models.SideSell
			}
			
			o := &models.Order{
				ID:        fmt.Sprintf("bench-%d-%d", time.Now().UnixNano(), i),
				Symbol:    "AAPL",
				Side:      side,
				Type:      models.TypeLimit,
				Price:     15000, 
				Quantity:  10,
				Timestamp: time.Now().UnixMilli(),
			}
			eng.SubmitOrder(o)
		}
	})
}