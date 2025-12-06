package engine

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/jaswantK19/order-matching-engine/internal/models"
)

func BenchmarkRealWorld(b *testing.B) {
	eng := NewEngine()
	eng.Start()
	
	symbols := []string{"AAPL", "GOOG", "TSLA", "AMZN", "MSFT"}
	
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			i++
			
			sym := symbols[rand.Intn(len(symbols))]
			
			side := models.SideBuy
			if i%2 == 0 {
				side = models.SideSell
			}
			
			price := int64(10000 + rand.Intn(100)) 
			
			qty := int64(1 + rand.Intn(100))

			o := &models.Order{
				ID:        fmt.Sprintf("bench-%d-%d", time.Now().UnixNano(), i),
				Symbol:    sym,
				Side:      side,
				Type:      models.TypeLimit,
				Price:     price,
				Quantity:  qty,
				Timestamp: time.Now().UnixMilli(),
			}
			eng.SubmitOrder(o)
		}
	})
}